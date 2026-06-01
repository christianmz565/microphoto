package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

type Client struct {
	minioClient *minio.Client
}

func NewClient(endpoint, accessKey, secretKey string, useSSL bool) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{minioClient: minioClient}, nil
}

// UploadObject uploads an object to the specified bucket and path
func (c *Client) UploadObject(ctx context.Context, bucket, path string, reader io.Reader, size int64, contentType string) (minio.UploadInfo, error) {
	return c.minioClient.PutObject(ctx, bucket, path, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

// DownloadObject downloads an object from the specified bucket and path
func (c *Client) DownloadObject(ctx context.Context, bucket, path string) (io.ReadCloser, error) {
	obj, err := c.minioClient.GetObject(ctx, bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	if _, err := obj.Stat(); err != nil {
		return nil, err
	}

	return obj, nil
}

// SetupLifecyclePolicy configures expiration rules for a bucket.
func (c *Client) SetupLifecyclePolicy(ctx context.Context, bucket string) error {
	config := lifecycle.NewConfiguration()
	config.Rules = []lifecycle.Rule{
		{
			ID:     "expire-partials",
			Status: "Enabled",
			Expiration: lifecycle.Expiration{
				Days: 1,
			},
			RuleFilter: lifecycle.Filter{
				Prefix: "partials/",
			},
		},
		{
			ID:     "expire-final",
			Status: "Enabled",
			Expiration: lifecycle.Expiration{
				Days: 3,
			},
			RuleFilter: lifecycle.Filter{
				Prefix: "final/",
			},
		},
	}

	return c.minioClient.SetBucketLifecycle(ctx, bucket, config)
}

// EnsureBucket ensures that the bucket exists
func (c *Client) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := c.minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		err = c.minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
