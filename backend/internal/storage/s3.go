package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/r3st/turnscore/config"
)

// S3Storage stores files in an S3-compatible object store (AWS S3, Cloudflare R2, MinIO).
type S3Storage struct {
	client *s3.Client
	bucket string
}

func newS3Client(cfg config.S3StorageConfig) *s3.Client {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	}
	awsCfg, _ := awsconfig.LoadDefaultConfig(context.Background(), opts...)

	s3Opts := []func(*s3.Options){
		func(o *s3.Options) { o.UsePathStyle = true },
	}
	if cfg.Endpoint != "" {
		endpoint := strings.TrimRight(cfg.Endpoint, "/")
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}
	return s3.NewFromConfig(awsCfg, s3Opts...)
}

func newS3Storage(cfg config.S3StorageConfig) (*S3Storage, error) {
	stor := &S3Storage{
		client: newS3Client(cfg),
		bucket: cfg.Bucket,
	}
	if err := EnsureBucketExists(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("s3 storage init: %w", err)
	}
	return stor, nil
}

// EnsureBucketExists creates the bucket if it does not already exist.
// Used for test setup and the MinIO init container equivalent in tests.
func EnsureBucketExists(ctx context.Context, cfg config.S3StorageConfig) error {
	client := newS3Client(cfg)
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		var bae *types.BucketAlreadyExists
		var bao *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bae) || errors.As(err, &bao) {
			return nil
		}
		return fmt.Errorf("create bucket: %w", err)
	}
	return nil
}

// Put uploads r to S3 under the given key.
func (s *S3Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("s3 put object: %w", err)
	}
	return nil
}

// Delete removes the object with the given key from S3.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete object: %w", err)
	}
	return nil
}

// Open returns a ReadCloser streaming the S3 object for the given key. Caller must close it.
func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get object: %w", err)
	}
	return out.Body, nil
}

// PublicURL returns the app-relative URL for the given key.
// Photos are served via the /uploads/* proxy handler, not directly from S3.
func (s *S3Storage) PublicURL(key string) string {
	return "/uploads/" + key
}
