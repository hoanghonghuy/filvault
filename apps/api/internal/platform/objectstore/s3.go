package objectstore

import (
	"context"
	"fmt"

	"filvault/internal/platform/config"

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
	client        *s3.Client
	presignClient *s3.Client
	bucket        string
}

func newS3(cfg config.Config) (*s3Store, error) {
	accessKey := cfg.S3AccessKey
	secretKey := cfg.S3SecretKey
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("S3 credentials are required")
	}

	client, err := buildS3Client(cfg.S3Endpoint, cfg.S3Region, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	// Presign against the public endpoint so the browser's Host header matches
	// the signed value. Fall back to the internal endpoint when unset.
	presignEndpoint := cfg.S3PublicEndpoint
	if presignEndpoint == "" {
		presignEndpoint = cfg.S3Endpoint
	}
	presignClient, err := buildS3Client(presignEndpoint, cfg.S3Region, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	return &s3Store{
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.S3Bucket,
	}, nil
}

func buildS3Client(endpoint, region, accessKey, secretKey string) (*s3.Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...any) (aws.Endpoint, error) {
		return aws.Endpoint{URL: endpoint, SigningRegion: region, HostnameImmutable: true}, nil
	})
	awsCfg, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion(region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awscfg.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}
