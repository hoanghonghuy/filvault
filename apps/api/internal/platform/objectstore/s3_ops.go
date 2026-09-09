package objectstore

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *s3Store) CheckReady(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("head bucket: %w", err)
	}
	return nil
}

func (s *s3Store) CreateUploadURL(ctx context.Context, key string, opts UploadOptions) (PresignedURL, error) {
	exp := opts.Expires
	if exp <= 0 {
		exp = 15 * time.Minute
	}
	presign := s3.NewPresignClient(s.presignClient)
	out, err := presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(opts.ContentType),
	}, s3.WithPresignExpires(exp))
	if err != nil {
		return PresignedURL{}, fmt.Errorf("presign upload: %w", err)
	}
	return PresignedURL{URL: out.URL, ExpiresAt: time.Now().UTC().Add(exp)}, nil
}

func (s *s3Store) CreateDownloadURL(ctx context.Context, key string, opts DownloadOptions) (PresignedURL, error) {
	exp := opts.Expires
	if exp <= 0 {
		exp = 5 * time.Minute
	}
	presign := s3.NewPresignClient(s.presignClient)
	out, err := presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(exp))
	if err != nil {
		return PresignedURL{}, fmt.Errorf("presign download: %w", err)
	}
	return PresignedURL{URL: out.URL, ExpiresAt: time.Now().UTC().Add(exp)}, nil
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
