package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ModuleVersionsBackend = &PostgresBackend{}

func (p *PostgresBackend) ModuleVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModuleVersionsRequest) (*models.ModuleVersionsResponse, error) {
	return nil, nil
}
