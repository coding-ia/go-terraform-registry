package sqlite_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProviderVersionsBackend = &SQLiteBackend{}

func (s *SQLiteBackend) ProviderVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionsRequest) (*models.ProviderVersionsResponse, error) {
	return nil, nil
}

func (s *SQLiteBackend) ProviderVersionPlatformsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionPlatformsRequest) (*models.ProviderVersionPlatformsResponse, error) {
	return nil, nil
}
