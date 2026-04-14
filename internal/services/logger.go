package services

import (
	"encoding/json"
	"log"
	"os"
	"pizza-son/internal/models"
	"sync"
	"time"
)

const logsDir = "data/logs"

type LogEntry struct {
	Username    string `json:"username"`
	Message     string `json:"message"`
	Timestamp   int64  `json:"timestamp"`
	Channel     string `json:"channel"`
	FirstMsg    bool   `json:"first_msg"`
	Subscriber  bool   `json:"subscriber"`
	VIP         bool   `json:"vip"`
	Mod         bool   `json:"mod"`
	Broadcaster bool   `json:"broadcaster"`
}

type LoggerService struct {
	mu       sync.Mutex
	files    map[string]*os.File
	encoders map[string]*json.Encoder
}

var LoggerServiceInstance *LoggerService

func NewLoggerService() {
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		log.Fatal("[Logger] Failed to create logs directory:", err)
	}
	LoggerServiceInstance = &LoggerService{
		files:    make(map[string]*os.File),
		encoders: make(map[string]*json.Encoder),
	}
	log.Println("[Logger] Service initialized")
}

func (s *LoggerService) getOrOpenFile(channel string) *json.Encoder {
	if enc, ok := s.encoders[channel]; ok {
		return enc
	}

	path := logsDir + "/" + channel + ".jsonl"

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[Logger] Failed to open log file for %s: %v", channel, err)
		return nil
	}

	s.files[channel] = f
	s.encoders[channel] = json.NewEncoder(f)
	return s.encoders[channel]
}

func (s *LoggerService) Log(msg models.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	enc := s.getOrOpenFile(msg.Channel)
	if enc == nil {
		return
	}

	entry := LogEntry{
		Username:    msg.User.Name,
		Message:     msg.Text,
		Timestamp:   time.Now().UnixMilli(),
		Channel:     msg.Channel,
		FirstMsg:    msg.FirstMessage,
		Subscriber:  msg.User.IsSubscriber,
		VIP:         msg.User.IsVIP,
		Mod:         msg.User.IsMod,
		Broadcaster: msg.User.IsBroadcaster,
	}

	if err := enc.Encode(entry); err != nil {
		log.Printf("[Logger] Failed to write log entry: %v", err)
	}
}

func (s *LoggerService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.files {
		f.Close()
	}
}
