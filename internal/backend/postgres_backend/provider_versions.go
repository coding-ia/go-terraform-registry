package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProviderVersionsBackend = &PostgresBackend{}

func (p *PostgresBackend) ProviderVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionsRequest) (*models.ProviderVersionsResponse, error) {
	return nil, nil
}

func (p *PostgresBackend) ProviderVersionPlatformsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionPlatformsRequest) (*models.ProviderVersionPlatformsResponse, error) {
	return nil, nil
}
