package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Store[T any] struct {
	path string
	data *T
}

func New[T any](path string, data *T) *Store[T] {
	return &Store[T]{path: path, data: data}
}

func (s *Store[T]) Path() string {
	return s.path
}

func (s *Store[T]) Data() *T {
	return s.data
}

func (s *Store[T]) Load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, s.data)
}

func (s *Store[T]) LoadIfExists() (bool, error) {
	err := s.Load()
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *Store[T]) Save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}

func (s *Store[T]) EnsureDir() error {
	return os.MkdirAll(filepath.Dir(s.path), 0755)
}
