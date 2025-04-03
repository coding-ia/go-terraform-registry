package storage

import (
	"context"
	"github.com/gin-gonic/gin"
)

type RegistryProviderStorage interface {
	ConfigureStorage(ctx context.Context) error
	GenerateUploadURL(ctx context.Context, path string) (string, error)
	GenerateDownloadURL(ctx context.Context, path string) (string, error)
}

type RegistryProviderStorageAssetEndpoint interface {
	ConfigureEndpoint(ctx context.Context, routerGroup *gin.RouterGroup)
}
