package s3_storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"log"
	"time"
)

var _ storage.RegistryProviderStorage = &S3Storage{}

type S3Storage struct {
	Config config.RegistryConfig

	client *s3.Client
}

func NewS3Storage(config config.RegistryConfig) storage.RegistryProviderStorage {
	return &S3Storage{
		Config: config,
	}
}

func (s *S3Storage) ConfigureStorage(ctx context.Context) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	if s.Config.AssumeRoleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		credentials := stscreds.NewAssumeRoleProvider(stsClient, s.Config.AssumeRoleARN)
		cfg.Credentials = aws.NewCredentialsCache(credentials)
	}

	s.client = s3.NewFromConfig(cfg)

	log.Println("Using S3 storage for providers & endpoints.")

	return nil
}

func (s *S3Storage) GenerateUploadURL(ctx context.Context, path string) (string, error) {
	preSignClient := s3.NewPresignClient(s.client)
	bucketName := s.Config.S3BucketName

	preSignedGetObject, err := preSignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
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
	bucketName := s.Config.S3BucketName

	preSignedGetObject, err := preSignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &path,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 60 * time.Minute
	})
	if err != nil {
		return "", err
	}
	return preSignedGetObject.URL, nil
}
