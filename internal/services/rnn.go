package services

import (
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"os"
)

const (
	rnnWeightsPath = "tools/lm/weights.json"
	rnnVocabPath   = "tools/lm/vocab.json"
)

type RNNService struct {
	char2idx  map[string]int
	idx2char  map[int]string
	vocabSize int
	weights   *GRUWeights
}

type GRUWeights struct {
	// Embedding
	Embed [][]float32

	// layer 1
	WIH0 [][]float32
	WHH0 [][]float32
	BIH0 []float32
	BHH0 []float32

	// layer 2
	WIH1 [][]float32
	WHH1 [][]float32
	BIH1 []float32
	BHH1 []float32

	// output
	WFC [][]float32
	BFC []float32
}

var RNNServiceInstance *RNNService

func NewRNNService() {
	vocab, err := loadVocab(rnnVocabPath)
	if err != nil {
		log.Fatal("[RNN] Failed to load vocab:", err)
	}

	weights, err := loadWeights(rnnWeightsPath)
	if err != nil {
		log.Fatal("[RNN] Failed to load weights:", err)
	}

	RNNServiceInstance = &RNNService{
		char2idx:  vocab.Char2Idx,
		idx2char:  vocab.Idx2Char,
		vocabSize: len(vocab.Char2Idx),
		weights:   weights,
	}
	log.Printf("[RNN] Service initialized, vocab size: %d", RNNServiceInstance.vocabSize)
}

type vocabData struct {
	Char2Idx map[string]int `json:"char2idx"`
	Idx2Char map[int]string `json:"-"`
}

func loadVocab(path string) (*vocabData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Char2Idx map[string]int    `json:"char2idx"`
		Idx2Char map[string]string `json:"idx2char"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	idx2char := make(map[int]string)
	for k, v := range raw.Idx2Char {
		var idx int
		json.Unmarshal([]byte(k), &idx)
		idx2char[idx] = v
	}
	return &vocabData{
		Char2Idx: raw.Char2Idx,
		Idx2Char: idx2char,
	}, nil
}

func loadWeights(path string) (*GRUWeights, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var w GRUWeights
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func sigmoid(x float32) float32 {
	return float32(1.0 / (1.0 + math.Exp(float64(-x))))
}

func tanh(x float32) float32 {
	return float32(math.Tanh(float64(x)))
}

func matVec(mat [][]float32, vec []float32) []float32 {
	out := make([]float32, len(mat))
	for i, row := range mat {
		var sum float32
		for j, v := range row {
			sum += v * vec[j]
		}
		out[i] = sum
	}
	return out
}

func addVecs(a, b []float32) []float32 {
	out := make([]float32, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func gruCell(x, h []float32, wih, whh [][]float32, bih, bhh []float32) []float32 {
	hiddenSize := len(h)

	ih := addVecs(matVec(wih, x), bih)
	hh := addVecs(matVec(whh, h), bhh)

	r := make([]float32, hiddenSize)
	z := make([]float32, hiddenSize)

	for i := range hiddenSize {
		r[i] = sigmoid(ih[i] + hh[i])
		z[i] = sigmoid(ih[i+hiddenSize] + hh[i+hiddenSize])
	}

	hNew := make([]float32, hiddenSize)
	for i := range hiddenSize {
		n := tanh(ih[i+2*hiddenSize] + r[i]*hh[i+2*hiddenSize])
		hNew[i] = (1-z[i])*n + z[i]*h[i]
	}
	return hNew
}

func (s *RNNService) Generate(seed string, length int, temperature float32) string {
	w := s.weights
	hiddenSize := len(w.WFC[0])

	h0 := make([]float32, hiddenSize)
	h1 := make([]float32, hiddenSize)

	result := []rune(seed)

	runStep := func(charIdx int) []float32 {
		embed := w.Embed[charIdx]
		h0 = gruCell(embed, h0, w.WIH0, w.WHH0, w.BIH0, w.BHH0)
		h1 = gruCell(h0, h1, w.WIH1, w.WHH1, w.BIH1, w.BHH1)
		logits := addVecs(matVec(w.WFC, h1), w.BFC)
		return logits
	}

	var logits []float32
	for _, c := range seed {
		idx, ok := s.char2idx[string(c)]
		if !ok {
			idx = 0
		}
		logits = runStep(idx)
	}

	// Generate
	for i := 0; i < length; i++ {
		idx := sampleLogits(logits, temperature)
		char, ok := s.idx2char[idx]
		if !ok || char == "\n" {
			break
		}
		result = append(result, []rune(char)...)
		logits = runStep(idx)
	}

	return string(result)
}

func (s *RNNService) RandomSeed() string {
	idx := rand.IntN(s.vocabSize)
	return s.idx2char[idx]
}

func sampleLogits(logits []float32, temperature float32) int {
	max := logits[0]
	for _, v := range logits {
		if v > max {
			max = v
		}
	}
	probs := make([]float64, len(logits))
	sum := 0.0
	for i, v := range logits {
		e := math.Exp(float64((v - max) / temperature))
		probs[i] = e
		sum += e
	}
	r := rand.Float64() * sum
	cumulative := 0.0
	for i, p := range probs {
		cumulative += p
		if r <= cumulative {
			return i
		}
	}

	return len(probs) - 1
}
