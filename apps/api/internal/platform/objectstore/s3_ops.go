package objectstore

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *s3Store) mapPublicURL(raw string) string {
	if s.publicEndpoint == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	pub, err := url.Parse(s.publicEndpoint)
	if err != nil {
		return raw
	}
	parsed.Scheme = pub.Scheme
	parsed.Host = pub.Host
	return parsed.String()
}

func (s *s3Store) CreateUploadURL(ctx context.Context, key string, opts UploadOptions) (PresignedURL, error) {
	exp := opts.Expires
	if exp <= 0 {
		exp = 15 * time.Minute
	}
	presign := s3.NewPresignClient(s.client)
	out, err := presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(opts.ContentType),
	}, s3.WithPresignExpires(exp))
	if err != nil {
		return PresignedURL{}, fmt.Errorf("presign upload: %w", err)
	}
	return PresignedURL{URL: s.mapPublicURL(out.URL), ExpiresAt: time.Now().UTC().Add(exp)}, nil
}

func (s *s3Store) CreateDownloadURL(ctx context.Context, key string, opts DownloadOptions) (PresignedURL, error) {
	exp := opts.Expires
	if exp <= 0 {
		exp = 5 * time.Minute
	}
	presign := s3.NewPresignClient(s.client)
	out, err := presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(exp))
	if err != nil {
		return PresignedURL{}, fmt.Errorf("presign download: %w", err)
	}
	return PresignedURL{URL: s.mapPublicURL(out.URL), ExpiresAt: time.Now().UTC().Add(exp)}, nil
}

func (s *s3Store) Head(ctx context.Context, key string) (ObjectStat, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectStat{}, mapS3NotFound(err)
	}
	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	ct := aws.ToString(out.ContentType)
	return ObjectStat{Size: size, ContentType: ct}, nil
}

func (s *s3Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func mapS3NotFound(err error) error {
	if err == nil {
		return nil
	}
	// Treat any Head failure as not found for Phase 1 complete flow.
	return ErrObjectNotFound
}
