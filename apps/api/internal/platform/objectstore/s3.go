package objectstore

import (
	"context"
	"fmt"

	"filnest/internal/platform/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewFromConfig builds S3-compatible storage (AWS S3 or MinIO).
func NewFromConfig(cfg config.Config) (ObjectStore, error) {
	if cfg.S3Endpoint == "" {
		return NewMemory(), nil
	}
	return newS3(cfg)
}

type s3Store struct {
	client *s3.Client
	bucket string
}

func newS3(cfg config.Config) (*s3Store, error) {
	accessKey := cfg.S3AccessKey
	secretKey := cfg.S3SecretKey
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...any) (aws.Endpoint, error) {
		return aws.Endpoint{URL: cfg.S3Endpoint, SigningRegion: cfg.S3Region, HostnameImmutable: true}, nil
	})
	awsCfg, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion(cfg.S3Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awscfg.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return &s3Store{client: client, bucket: cfg.S3Bucket}, nil
}
