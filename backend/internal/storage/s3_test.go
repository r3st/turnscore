package storage_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/storage"
)

func TestS3Storage_PutDeletePublicURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MinIO integration test in short mode")
	}
	ctx := context.Background()

	const (
		minioUser   = "minioadmin"
		minioPass   = "minioadmin"
		minioBucket = "turnscore-test"
		minioRegion = "us-east-1"
	)

	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     minioUser,
			"MINIO_ROOT_PASSWORD": minioPass,
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
	}

	minioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = minioC.Terminate(ctx) })

	host, err := minioC.Host(ctx)
	require.NoError(t, err)
	port, err := minioC.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)
	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Create bucket via S3 API (MinIO starts without buckets).
	adminCfg := config.StorageConfig{
		Backend: "s3",
		S3: config.S3StorageConfig{
			Endpoint:  endpoint,
			Bucket:    minioBucket,
			Region:    minioRegion,
			AccessKey: minioUser,
			SecretKey: minioPass,
		},
	}
	adminStore, err := storage.New(adminCfg)
	require.NoError(t, err)

	// Put a small file — this implicitly tests bucket creation when using path-style.
	// We need to create the bucket first via the S3 CreateBucket call.
	err = storage.EnsureBucketExists(ctx, adminCfg.S3)
	require.NoError(t, err)

	s, err := storage.New(adminCfg)
	require.NoError(t, err)

	t.Run("Put and PublicURL", func(t *testing.T) {
		key := "tables/abc/general/photo.jpg"
		content := "fake-jpeg-content"

		err := s.Put(ctx, key, strings.NewReader(content), int64(len(content)), "image/jpeg")
		require.NoError(t, err)

		url := s.PublicURL(key)
		assert.Equal(t, fmt.Sprintf("%s/%s/%s", endpoint, minioBucket, key), url)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "tables/abc/general/to-delete.jpg"
		err := s.Put(ctx, key, strings.NewReader("data"), 4, "image/jpeg")
		require.NoError(t, err)

		err = s.Delete(ctx, key)
		require.NoError(t, err)

		// Deleting a non-existent key must not error (idempotent).
		err = s.Delete(ctx, key)
		assert.NoError(t, err)
	})

	_ = adminStore // used only for setup
}
