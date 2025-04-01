package s3_storage

import (
	"context"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
)

var _ storage.RegistryProviderStorage = &S3Storage{}

type S3Storage struct {
	S3BucketName   string
	S3BucketRegion string
}

func NewS3Storage(config config.RegistryConfig) storage.RegistryProviderStorage {
	return &S3Storage{
		S3BucketName:   config.S3BucketName,
		S3BucketRegion: config.S3BucketRegion,
	}
}

func (s S3Storage) ConfigureStorage(ctx context.Context) {

}
