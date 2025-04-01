package s3_storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"time"
)

var _ storage.RegistryProviderStorage = &S3Storage{}

type S3Storage struct {
	S3BucketName   string
	S3BucketRegion string

	s3svc *s3.S3
}

func NewS3Storage(config config.RegistryConfig) storage.RegistryProviderStorage {
	return &S3Storage{
		S3BucketName:   config.S3BucketName,
		S3BucketRegion: config.S3BucketRegion,
	}
}

func (s *S3Storage) ConfigureStorage(ctx context.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.S3BucketRegion)},
	)
	if err != nil {
		return err
	}

	s.s3svc = s3.New(sess)
	return nil
}

func (s *S3Storage) GenerateUploadURL(ctx context.Context, path string) (string, error) {
	req, _ := s.s3svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.S3BucketName),
		Key:    aws.String(path),
	})
	url, err := req.Presign(60 * time.Minute)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *S3Storage) GenerateDownloadURL(ctx context.Context, path string) (string, error) {
	return "", nil
}
