package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/storage"
)

func TestLocalStorage_PutAndPublicURL(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(config.StorageConfig{
		Backend: "local",
		Local: config.LocalStorageConfig{
			Path:    dir,
			BaseURL: "http://localhost:8080",
		},
	})
	require.NoError(t, err)

	content := "hello turnscore"
	key := "tables/test-table/general/photo.jpg"

	err = s.Put(context.Background(), key, strings.NewReader(content), int64(len(content)), "image/jpeg")
	require.NoError(t, err)

	url := s.PublicURL(key)
	assert.Equal(t, "http://localhost:8080/uploads/"+key, url)
}

func TestLocalStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(config.StorageConfig{
		Backend: "local",
		Local:   config.LocalStorageConfig{Path: dir, BaseURL: "http://localhost:8080"},
	})
	require.NoError(t, err)

	key := "tables/test-table/general/to-delete.jpg"
	err = s.Put(context.Background(), key, strings.NewReader("data"), 4, "image/jpeg")
	require.NoError(t, err)

	err = s.Delete(context.Background(), key)
	require.NoError(t, err)

	// Deleting again must not return an error (idempotent)
	err = s.Delete(context.Background(), key)
	assert.NoError(t, err)
}

func TestLocalStorage_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()
	s, err := storage.New(config.StorageConfig{
		Backend: "local",
		Local:   config.LocalStorageConfig{Path: dir, BaseURL: "http://localhost:8080"},
	})
	require.NoError(t, err)

	key := "deep/nested/path/file.jpg"
	err = s.Put(context.Background(), key, strings.NewReader("content"), 7, "image/jpeg")
	assert.NoError(t, err)
}
