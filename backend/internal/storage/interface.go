// Package storage provides an interface-based file storage abstraction.
// The backend (local filesystem or S3-compatible) is selected via configuration.
package storage

import (
	"context"
	"io"

	"github.com/r3st/turnscore/config"
)

// Storage defines the operations for storing and retrieving files.
type Storage interface {
	// Put stores the content from r under the given key.
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	// Delete removes the file identified by key.
	Delete(ctx context.Context, key string) error
	// PublicURL returns the publicly accessible URL for the given key.
	PublicURL(key string) string
}

// New creates a Storage implementation based on the provided configuration.
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Backend {
	case "s3":
		return newS3Storage(cfg.S3)
	default: // "local"
		return newLocalStorage(cfg.Local)
	}
}
