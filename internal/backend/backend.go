package backend

import (
	"context"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
)

type RegistryProviderBackend interface {
	ConfigureBackend(ctx context.Context)
	GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters) (*models.TerraformProviderPlatformResponse, error)
	GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters) (*models.TerraformAvailableProvider, error)
}
