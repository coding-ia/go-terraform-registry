package s3_storage

import (
	"context"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"time"
)

var _ storage.RegistryProviderStorage = &S3Storage{}

type S3Storage struct {
	S3BucketName   string
	S3BucketRegion string

	client *s3.Client
}

func NewS3Storage(config config.RegistryConfig) storage.RegistryProviderStorage {
	return &S3Storage{
		S3BucketName:   config.S3BucketName,
		S3BucketRegion: config.S3BucketRegion,
	}
}

func (s *S3Storage) ConfigureStorage(ctx context.Context) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	s.client = s3.NewFromConfig(cfg)
	return nil
}

func (s *S3Storage) GenerateUploadURL(ctx context.Context, path string) (string, error) {
	preSignClient := s3.NewPresignClient(s.client)
	preSignedGetObject, err := preSignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.S3BucketName,
		Key:    &path,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 60 * time.Minute
	})
	if err != nil {
		return "", err
	}
	return preSignedGetObject.URL, nil
}

func (s *S3Storage) GenerateDownloadURL(ctx context.Context, path string) (string, error) {
	preSignClient := s3.NewPresignClient(s.client)
	preSignedGetObject, err := preSignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.S3BucketName,
		Key:    &path,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 60 * time.Minute
	})
	if err != nil {
		return "", err
	}
	return preSignedGetObject.URL, nil
}
