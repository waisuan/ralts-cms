// Package s3 provides interfaces and types for AWS S3 client operations.
package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client is the interface for the S3 client
//
//go:generate mockgen -destination=./mock_s3_client.go -package=s3 -source=client.go
type Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
}
