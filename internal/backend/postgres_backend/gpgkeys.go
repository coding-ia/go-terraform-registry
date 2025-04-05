package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
)

var _ backend.GPGKeysBackend = &PostgresBackend{}

func (p *PostgresBackend) GPGKeysAdd(ctx context.Context, request models.GPGKeysRequest) (*models.GPGKeysResponse, error) {
	return nil, nil
}
