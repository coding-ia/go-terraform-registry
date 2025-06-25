package storage

import (
	"context"
	"github.com/go-chi/chi/v5"
)

type RegistryProviderStorage interface {
	ConfigureStorage(ctx context.Context) error
	GenerateUploadURL(ctx context.Context, path string) (string, error)
	GenerateDownloadURL(ctx context.Context, path string) (string, error)
	RemoveFile(ctx context.Context, path string) error
	RemoveDirectory(ctx context.Context, path string) error
}

type RegistryProviderStorageAssetEndpoint interface {
	ConfigureEndpoint(ctx context.Context, cr *chi.Mux)
}
