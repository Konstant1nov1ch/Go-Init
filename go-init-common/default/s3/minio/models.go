package minio

import (
	"time"

	"github.com/google/uuid"
)

// ObjectMetadata represents the metadata of an object stored in S3-compatible storage
type ObjectMetadata struct {
	// ID is a unique identifier for the object metadata
	ID string `json:"id"`

	// BucketName is the name of the bucket the object is stored in
	BucketName string `json:"bucketName"`

	// ObjectName is the name of the object in the bucket
	ObjectName string `json:"objectName"`

	// Size is the size of the object in bytes
	Size int64 `json:"objectSize"`

	// ContentType is the MIME type of the object
	ContentType string `json:"contentType"`

	// ETag is the entity tag of the object
	ETag string `json:"etag"`

	// CreatedAt is the timestamp when the object was created
	CreatedAt time.Time `json:"createdAt"`

	// LastModified is the timestamp when the object was last modified
	LastModified time.Time `json:"lastModified"`

	// AdditionalMetadata contains any additional metadata for the object
	AdditionalMetadata map[string]string `json:"additionalMetadata,omitempty"`
}

// ArchiveMetadata represents metadata specific to an archive object
type ArchiveMetadata struct {
	// ID is a unique identifier used to correlate this archive with a request
	ID string `json:"id,omitempty"`

	// ObjectMetadata contains the base object metadata
	ObjectMetadata

	// ArchiveType indicates the type of the archive (e.g., "zip", "tar", etc.)
	ArchiveType string `json:"archiveType"`

	// PresignedURL is an optional field for storing a pre-signed URL for accessing the archive
	PresignedURL string `json:"presignedURL,omitempty"`

	// ExpiresAt is an optional field for storing the expiration time of the presigned URL
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// NewObjectMetadata creates a new ObjectMetadata instance
func NewObjectMetadata(bucketName, objectName string, size int64, contentType, etag string) ObjectMetadata {
	now := time.Now()
	return ObjectMetadata{
		ID:                 uuid.New().String(),
		BucketName:         bucketName,
		ObjectName:         objectName,
		Size:               size,
		ContentType:        contentType,
		ETag:               etag,
		CreatedAt:          now,
		LastModified:       now,
		AdditionalMetadata: make(map[string]string),
	}
}

// NewArchiveMetadata creates a new ArchiveMetadata instance
func NewArchiveMetadata(objectMetadata ObjectMetadata, archiveType string) ArchiveMetadata {
	return ArchiveMetadata{
		ID:             objectMetadata.ID,
		ObjectMetadata: objectMetadata,
		ArchiveType:    archiveType,
	}
}

// SetPresignedURL sets the presigned URL and its expiration time
func (am *ArchiveMetadata) SetPresignedURL(url string, expiresIn time.Duration) {
	am.PresignedURL = url
	expiresAt := time.Now().Add(expiresIn)
	am.ExpiresAt = &expiresAt
}

// IsPresignedURLValid checks if the presigned URL is still valid
func (am *ArchiveMetadata) IsPresignedURLValid() bool {
	if am.ExpiresAt == nil {
		return false
	}
	return time.Now().Before(*am.ExpiresAt)
}
