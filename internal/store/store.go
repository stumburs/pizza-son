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

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write to a temporary file first so a crash cannot leave a half-written
	// JSON file behind.
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(s.path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

func (s *Store[T]) EnsureDir() error {
	return os.MkdirAll(filepath.Dir(s.path), 0755)
}
