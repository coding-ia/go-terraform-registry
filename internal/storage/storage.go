package storage

import "context"

type RegistryProviderStorage interface {
	ConfigureStorage(ctx context.Context) error
	GenerateUploadURL(ctx context.Context, path string) (string, error)
	GenerateDownloadURL(ctx context.Context, path string) (string, error)
}
