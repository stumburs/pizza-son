package services

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"strings"
	"sync"
	"time"
)

type SayFunc func(channel, message string)

type STTService struct {
	mu                   sync.Mutex
	sayFn                SayFunc
	enabled              map[string]bool
	quit                 chan struct{}
	streamlinkPath       string
	ffmpegPath           string
	whisperPath          string
	whisperModel         string
	captureInterval      time.Duration
	streamlinkAvailable  bool
	ffmpegAvailable      bool
	whisperAvailable     bool
}

var STTServiceInstance *STTService

func NewSTTService(sayFn SayFunc) {
	cfg := config.Get()

	var interval time.Duration
	if cfg.STT.CaptureInterval <= 0 {
		interval = 30 * time.Second
	} else {
		interval = time.Duration(cfg.STT.CaptureInterval) * time.Second
	}

	svc := &STTService{
		sayFn:           sayFn,
		enabled:         make(map[string]bool),
		quit:            make(chan struct{}),
		streamlinkPath:  cfg.STT.StreamlinkPath,
		ffmpegPath:      cfg.STT.FFmpegPath,
		whisperPath:     cfg.STT.WhisperPath,
		whisperModel:    cfg.STT.WhisperModel,
		captureInterval: interval,
	}

	if cfg.STT.Enabled {
		svc.checkDependencies()
		go svc.captureLoop()
		log.Println("[STT] Service started")
	} else {
		log.Println("[STT] Service disabled in config")
	}

	STTServiceInstance = svc
}

func (s *STTService) checkDependencies() {
	if _, err := exec.LookPath(s.streamlinkPath); err != nil {
		log.Printf("[STT] streamlink not found at '%s' - audio capture disabled", s.streamlinkPath)
	} else {
		s.streamlinkAvailable = true
		log.Printf("[STT] streamlink found at '%s'", s.streamlinkPath)
	}

	if _, err := exec.LookPath(s.ffmpegPath); err != nil {
		log.Printf("[STT] ffmpeg not found at '%s' - audio capture disabled", s.ffmpegPath)
	} else {
		s.ffmpegAvailable = true
		log.Printf("[STT] ffmpeg found at '%s'", s.ffmpegPath)
	}

	if s.whisperModel != "" {
		if _, err := os.Stat(s.whisperModel); err != nil {
			log.Printf("[STT] whisper model not found at '%s'", s.whisperModel)
		} else {
			if _, err := exec.LookPath(s.whisperPath); err != nil {
				log.Printf("[STT] whisper binary not found at '%s'", s.whisperPath)
			} else {
				s.whisperAvailable = true
				log.Printf("[STT] whisper found at '%s' with model '%s'", s.whisperPath, s.whisperModel)
			}
		}
	} else {
		log.Println("[STT] No whisper model configured - STT disabled")
	}

	if !s.streamlinkAvailable || !s.ffmpegAvailable {
		log.Println("[STT] Missing streamlink or ffmpeg - audio capture will not work")
	}

	if !s.whisperAvailable {
		log.Println("[STT] Missing whisper binary or model - transcription will not work")
	}
}

func (s *STTService) captureLoop() {
	ticker := time.NewTicker(s.captureInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			s.processEnabledChannels()
		}
	}
}

func (s *STTService) processEnabledChannels() {
	s.mu.Lock()
	channels := make([]string, 0, len(s.enabled))
	for ch := range s.enabled {
		channels = append(channels, ch)
	}
	s.mu.Unlock()

	log.Printf("[STT] Processing %d enabled channels: %v", len(channels), channels)

	for _, channel := range channels {
		if !s.streamlinkAvailable {
			log.Printf("[STT] %s: streamlink not available, skipping", channel)
			return
		}
		if !s.ffmpegAvailable {
			log.Printf("[STT] %s: ffmpeg not available, skipping", channel)
			return
		}
		if !s.whisperAvailable {
			log.Printf("[STT] %s: whisper not available, skipping", channel)
			return
		}

		info := TwitchServiceInstance.GetStreamInfo(channel)
		log.Printf("[STT] %s: viewer_count=%d", channel, info.ViewerCount)
		if info.ViewerCount <= 0 {
			log.Printf("[STT] %s: stream offline or no viewer data, skipping", channel)
			continue
		}

		text, err := s.captureAndTranscribe(channel)
		if err != nil {
			log.Printf("[STT] %s: capture failed: %v", channel, err)
			continue
		}

		log.Printf("[STT] %s: transcribed %d chars", channel, len(text))
		if text == "" {
			continue
		}

		s.handleTranscript(channel, text)
	}
}

func (s *STTService) captureAndTranscribe(channel string) (string, error) {
	if !s.streamlinkAvailable || !s.ffmpegAvailable {
		return "", fmt.Errorf("streamlink or ffmpeg not available")
	}

	streamURL, err := s.getStreamURL(channel)
	if err != nil {
		return "", fmt.Errorf("get stream URL: %w", err)
	}

	captureFile, err := s.captureAudio(streamURL, channel)
	if err != nil {
		return "", fmt.Errorf("capture audio: %w", err)
	}
	defer os.Remove(captureFile)

	fi, fiErr := os.Stat(captureFile)
	if fiErr == nil {
		log.Printf("[STT] %s: captured audio %s (%d bytes)", channel, captureFile, fi.Size())
	}

	if !s.whisperAvailable {
		return "", fmt.Errorf("whisper not available")
	}

	text, err := s.transcribe(captureFile)
	if err != nil {
		return "", fmt.Errorf("transcribe: %w", err)
	}

	return text, nil
}

func (s *STTService) getStreamURL(channel string) (string, error) {
	cmd := exec.Command(s.streamlinkPath, "--stream-url", fmt.Sprintf("twitch.tv/%s", channel), "best")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("streamlink failed: %w", err)
	}
	url := strings.TrimSpace(out.String())
	if url == "" {
		return "", fmt.Errorf("empty stream URL")
	}
	return url, nil
}

func (s *STTService) captureAudio(streamURL, channel string) (string, error) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.wav", channel, time.Now().Unix()))
	cmd := exec.Command(s.ffmpegPath,
		"-y",
		"-i", streamURL,
		"-t", fmt.Sprintf("%.0f", s.captureInterval.Seconds()),
		"-ac", "1",
		"-ar", "16000",
		"-f", "wav",
		tmpFile,
	)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}
	return tmpFile, nil
}

func (s *STTService) transcribe(audioFile string) (string, error) {
	txtFile := audioFile + ".txt"
	cmd := exec.Command(s.whisperPath,
		"-m", s.whisperModel,
		"-f", audioFile,
		"-otxt",
		"--language", "en",
	)
	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("[STT] whisper error: %v, stderr: %s", err, stderr.String()[:min(stderr.Len(), 500)])
		return "", fmt.Errorf("whisper failed: %w", err)
	}

	if stderr.Len() > 0 {
		log.Printf("[STT] whisper stderr (%d bytes): %s", stderr.Len(), stderr.String()[:min(stderr.Len(), 300)])
	}

	data, err := os.ReadFile(txtFile)
	if err != nil {
		// Fallback: try base filename (whisper may strip extension)
		base := strings.TrimSuffix(audioFile, ".wav")
		if data2, err2 := os.ReadFile(base + ".txt"); err2 == nil {
			data = data2
			txtFile = base + ".txt"
		} else {
			log.Printf("[STT] whisper output file not found: %s or %s", txtFile, base+".txt")
			return "", nil
		}
	}
	_ = os.Remove(txtFile)

	log.Printf("[STT] whisper output (%d bytes): %q", len(data), string(data[:min(len(data), 500)]))
	parsed := s.parseWhisperOutput(string(data))
	log.Printf("[STT] whisper parsed: %q", parsed)
	return parsed, nil
}

func (s *STTService) parseWhisperOutput(output string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		idx := strings.Index(line, "]")
		if idx >= 0 {
			line = strings.TrimSpace(line[idx+1:])
		}
		if len([]rune(line)) < 3 {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, " ")
}

func (s *STTService) handleTranscript(channel, transcript string) {
	if len(strings.Fields(transcript)) < 3 {
		return
	}

	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return
	}

	ambient := getRecentAmbient(channel, 10)

	prompt := fmt.Sprintf(
		"Recent chat:\n%s\n\nThe streamer just said: \"%s\"",
		ambient,
		transcript,
	)

	msg := models.Message{
		Channel:  channel,
		Text:     "!streamlisten",
		Platform: models.PlatformTwitch,
		User: models.MessageUser{
			Name:        "StreamListener",
			DisplayName: "StreamListener",
		},
	}
	res, err := OllamaServiceInstance.GenerateChatResponse(msg,
		WithPrompt(prompt),
		WithCommand("streamlisten"),
	)

	if err != nil {
		log.Printf("[STT] %s: ollama error: %v", channel, err)
		return
	}

	response := ""
	if res.Message.Content != nil {
		response = strings.TrimSpace(*res.Message.Content)
	}

	if response == "" || response == "SILENCE" {
		log.Printf("[STT] %s: LLM returned SILENCE, not responding", channel)
		return
	}

	log.Printf("[STT] %s: LLM responded: %s", channel, response)
	if s.sayFn != nil {
		s.sayFn(channel, response)
	}
}

func getRecentAmbient(channel string, count int) string {
	return OllamaServiceInstance.GetRecentAmbient(channel, count)
}

func (s *STTService) TestTranscript(channel, text string) {
	if s.sayFn != nil {
		s.sayFn(channel, fmt.Sprintf("[STT test] Heard: \"%s\"", text))
	}
	s.handleTranscript(channel, text)
}

func (s *STTService) IsEnabled(channel string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled[channel]
}

func (s *STTService) SetEnabled(channel string, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if enabled {
		s.enabled[channel] = true
	} else {
		delete(s.enabled, channel)
	}
}

func (s *STTService) Stop() {
	close(s.quit)
}
