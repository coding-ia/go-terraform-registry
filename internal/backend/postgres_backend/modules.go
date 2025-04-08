package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ModulesBackend = &PostgresBackend{}

func (p *PostgresBackend) ModulesCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModulesRequest) (*models.ModulesResponse, error) {
	module := &Module{
		Name:         request.Data.Attributes.Name,
		Namespace:    request.Data.Attributes.Namespace,
		Organization: parameters.Organization,
		RegistryName: request.Data.Attributes.RegistryName,
		Provider:     request.Data.Attributes.Provider,
	}

	err := modulesInsert(ctx, p.db, module)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *PostgresBackend) ModulesGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ModulesResponse, error) {
	return nil, nil
}
