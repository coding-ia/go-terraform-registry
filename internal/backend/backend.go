package backend

import (
	"context"
	apimodels "go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
)

type RegistryProviderBackend interface {
	ConfigureBackend(ctx context.Context) error
	GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error)
	GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters, userParameters registrytypes.UserParameters) (*models.TerraformAvailableProvider, error)
	GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error)
	GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error)

	ProvidersCreate(ctx context.Context, parameters registrytypes.APIParameters, request apimodels.ProvidersRequest) (*apimodels.ProvidersResponse, error)
	GPGKeysAdd(ctx context.Context, request apimodels.GPGKeysRequest) (*apimodels.GPGKeysResponse, error)
	RegistryProviderVersions(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionsRequest) (*models.RegistryProviderVersionsResponse, error)
	RegistryProviderVersionPlatforms(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionPlatformsRequest) (*models.RegistryProviderVersionPlatformsResponse, error)
}
