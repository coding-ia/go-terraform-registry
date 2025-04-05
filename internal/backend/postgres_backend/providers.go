package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProvidersBackend = &PostgresBackend{}

func (p *PostgresBackend) ProvidersCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProvidersRequest) (*models.ProvidersResponse, error) {
	return nil, nil
}
