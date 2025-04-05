package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.RegistryBackend = &PostgresBackend{}

func (p *PostgresBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error) {
	return nil, nil
}

func (p *PostgresBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters, userParameters registrytypes.UserParameters) (*models.TerraformAvailableProvider, error) {
	return nil, nil
}

func (p *PostgresBackend) GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	return nil, nil
}

func (p *PostgresBackend) GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
	return nil, nil
}
