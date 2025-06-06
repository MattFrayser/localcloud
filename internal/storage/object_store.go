package storage

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStore struct {
	client    *minio.Client
	buckets   map[string]bool
	mu        sync.RWMutex
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
}

func NewObjectStore(endpoint, accessKey, secretKey string, useSSL bool) (*ObjectStore, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	return &ObjectStore{
		client:    client,
		buckets:   make(map[string]bool),
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
		useSSL:    useSSL,
	}, nil
}

func (os *ObjectStore) PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	// Ensure bucket exists
	if !os.buckets[bucket] {
		err := os.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			// Check if bucket already exists
			exists, err := os.client.BucketExists(ctx, bucket)
			if err != nil || !exists {
				return fmt.Errorf("failed to create bucket: %v", err)
			}
		}
		os.buckets[bucket] = true
	}

	// Upload the object
	_, err := os.client.PutObject(ctx, bucket, key, data, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload object: %v", err)
	}

	return nil
}

func (os *ObjectStore) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	if !os.buckets[bucket] {
		exists, err := os.client.BucketExists(ctx, bucket)
		if err != nil || !exists {
			return nil, fmt.Errorf("bucket does not exist: %s", bucket)
		}
		os.buckets[bucket] = true
	}

	object, err := os.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}

	return object, nil
}

func (os *ObjectStore) DeleteObject(ctx context.Context, bucket, key string) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	if !os.buckets[bucket] {
		exists, err := os.client.BucketExists(ctx, bucket)
		if err != nil || !exists {
			return fmt.Errorf("bucket does not exist: %s", bucket)
		}
		os.buckets[bucket] = true
	}

	err := os.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

func (os *ObjectStore) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	os.mu.RLock()
	defer os.mu.RUnlock()

	if !os.buckets[bucket] {
		exists, err := os.client.BucketExists(ctx, bucket)
		if err != nil || !exists {
			return nil, fmt.Errorf("bucket does not exist: %s", bucket)
		}
		os.buckets[bucket] = true
	}

	var objects []string
	objectCh := os.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %v", object.Err)
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}
