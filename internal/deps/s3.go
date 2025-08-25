package deps

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client creates a new AWS S3 client using the configuration from Config.
// Supports both real AWS and LocalStack endpoints.
func NewS3Client(ctx context.Context, cfg *Config) (*s3.Client, error) {
	var awsCfg aws.Config
	var err error

	// Configure options for AWS SDK
	configOptions := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(cfg.AWSDefaultRegion),
	}

	// If we have explicit credentials, use them
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		configOptions = append(configOptions, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKeyID,
				cfg.AWSSecretAccessKey,
				"", // session token (empty for static credentials)
			),
		))
	}

	awsCfg, err = awsConfig.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with additional options for LocalStack compatibility
	s3Options := []func(*s3.Options){}

	// If endpoint URL is specified (e.g., for LocalStack), use it
	if cfg.AWSEndpointURL != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.AWSEndpointURL)
		})
	}

	// Force path-style addressing if specified (required for LocalStack)
	if cfg.AWSS3ForcePathStyle {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	return s3.NewFromConfig(awsCfg, s3Options...), nil
}
