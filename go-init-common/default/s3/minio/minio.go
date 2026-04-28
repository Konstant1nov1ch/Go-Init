package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gitlab.com/go-init/go-init-common/default/logger"
)

// Config contains MinIO connection parameters
type Config struct {
	Endpoint        string `yaml:"endpoint" env:"MINIO_ENDPOINT" default:"localhost:9000"`
	AccessKeyID     string `yaml:"access_key_id" env:"MINIO_ACCESS_KEY_ID"`
	SecretAccessKey string `yaml:"secret_access_key" env:"MINIO_SECRET_ACCESS_KEY"`
	UseSSL          bool   `yaml:"use_ssl" env:"MINIO_USE_SSL" default:"false"`
	Region          string `yaml:"region" env:"MINIO_REGION" default:""`
	DefaultBucket   string `yaml:"default_bucket" env:"MINIO_DEFAULT_BUCKET" default:"go-init-archives"`
}

// Client is a wrapper around the MinIO client
type Client struct {
	client *minio.Client
	config *Config
	log    *logger.Logger
}

// New creates a new MinIO client
func New(config *Config, log *logger.Logger) (*Client, error) {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	log.Info(fmt.Sprintf("MinIO client initialized for endpoint: %s", config.Endpoint))
	return &Client{
		client: minioClient,
		config: config,
		log:    log,
	}, nil
}

// EnsureBucketExists checks if a bucket exists and creates it if it doesn't
func (c *Client) EnsureBucketExists(ctx context.Context, bucketName string) error {
	// If no bucket name is provided, use the default bucket
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		c.log.Info(fmt.Sprintf("Creating bucket: %s", bucketName))
		err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region: c.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
		}
		c.log.Info(fmt.Sprintf("Bucket %s created successfully", bucketName))
	} else {
		c.log.Info(fmt.Sprintf("Bucket %s already exists", bucketName))
	}

	return nil
}

// UploadObject uploads an object to MinIO with the provided data
func (c *Client) UploadObject(ctx context.Context, bucketName, objectName string, data []byte, contentType string) (minio.UploadInfo, error) {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	// Ensure the bucket exists
	if err := c.EnsureBucketExists(ctx, bucketName); err != nil {
		return minio.UploadInfo{}, err
	}

	reader := bytes.NewReader(data)
	objectSize := int64(len(data))

	// If content type is not provided, try to determine it from the file extension
	if contentType == "" {
		contentType = getContentType(objectName)
	}

	c.log.Info(fmt.Sprintf("Uploading object %s to bucket %s (size: %d bytes)", objectName, bucketName, objectSize))
	info, err := c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to upload object %s: %w", objectName, err)
	}

	c.log.Info(fmt.Sprintf("Object %s uploaded successfully, ETag: %s", objectName, info.ETag))
	return info, nil
}

// UploadObjectFromReader uploads an object to MinIO from a reader
func (c *Client) UploadObjectFromReader(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (minio.UploadInfo, error) {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	// Ensure the bucket exists
	if err := c.EnsureBucketExists(ctx, bucketName); err != nil {
		return minio.UploadInfo{}, err
	}

	// If content type is not provided, try to determine it from the file extension
	if contentType == "" {
		contentType = getContentType(objectName)
	}

	c.log.Info(fmt.Sprintf("Uploading object %s to bucket %s (size: %d bytes)", objectName, bucketName, objectSize))
	info, err := c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to upload object %s: %w", objectName, err)
	}

	c.log.Info(fmt.Sprintf("Object %s uploaded successfully, ETag: %s", objectName, info.ETag))
	return info, nil
}

// DownloadObject downloads an object from MinIO
func (c *Client) DownloadObject(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	c.log.Info(fmt.Sprintf("Downloading object %s from bucket %s", objectName, bucketName))
	obj, err := c.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", objectName, err)
	}
	defer obj.Close()

	objInfo, err := obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get object stats: %w", err)
	}

	buffer := make([]byte, objInfo.Size)
	_, err = obj.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	c.log.Info(fmt.Sprintf("Object %s downloaded successfully (size: %d bytes)", objectName, objInfo.Size))
	return buffer, nil
}

// DeleteObject deletes an object from MinIO
func (c *Client) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	c.log.Info(fmt.Sprintf("Deleting object %s from bucket %s", objectName, bucketName))
	err := c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", objectName, err)
	}

	c.log.Info(fmt.Sprintf("Object %s deleted successfully", objectName))
	return nil
}

// GeneratePresignedURL generates a presigned URL for accessing an object
func (c *Client) GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (*url.URL, error) {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	c.log.Info(fmt.Sprintf("Generating presigned URL for object %s in bucket %s (expiry: %s)", objectName, bucketName, expiry))
	presignedURL, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	c.log.Info(fmt.Sprintf("Presigned URL generated successfully: %s", presignedURL.String()))
	return presignedURL, nil
}

// ListObjects lists all objects in a bucket with an optional prefix
func (c *Client) ListObjects(ctx context.Context, bucketName, prefix string) ([]minio.ObjectInfo, error) {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	c.log.Info(fmt.Sprintf("Listing objects in bucket %s with prefix %s", bucketName, prefix))

	var objects []minio.ObjectInfo
	objectCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		objects = append(objects, object)
	}

	c.log.Info(fmt.Sprintf("Found %d objects in bucket %s with prefix %s", len(objects), bucketName, prefix))
	return objects, nil
}

// ObjectExists checks if an object exists in a bucket
func (c *Client) ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	// Use default bucket if not specified
	if bucketName == "" {
		bucketName = c.config.DefaultBucket
	}

	_, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// If the error is about the object not existing, return false without an error
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if object exists: %w", err)
	}

	return true, nil
}

// getContentType returns a MIME type based on file extension
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".zip":
		return "application/zip"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}
