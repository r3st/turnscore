package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/r3st/turnscore/config"
)

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	basePath string
	baseURL  string
}

func newLocalStorage(cfg config.LocalStorageConfig) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		return nil, fmt.Errorf("creating local storage directory: %w", err)
	}
	return &LocalStorage{
		basePath: cfg.Path,
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
	}, nil
}

// Put writes the content of r to basePath/key, creating intermediate directories.
func (s *LocalStorage) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	fullPath := filepath.Join(s.basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	if _, err = io.Copy(f, r); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

// Delete removes the file at basePath/key.
func (s *LocalStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, filepath.FromSlash(key))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting file: %w", err)
	}
	return nil
}

// PublicURL returns the full URL for accessing the file.
func (s *LocalStorage) PublicURL(key string) string {
	return s.baseURL + "/uploads/" + key
}
