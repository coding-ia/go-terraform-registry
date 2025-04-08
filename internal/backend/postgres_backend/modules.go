package postgres_backend

import (
	"context"
	"fmt"
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

	resp := &models.ModulesResponse{
		Data: models.ModulesDataResponse{
			ID:   module.ID,
			Type: "registry-modules",
			Attributes: models.ModulesAttributesResponse{
				Name:         module.Name,
				Namespace:    module.Namespace,
				RegistryName: module.RegistryName,
				Provider:     module.Provider,
			},
		},
	}

	return resp, nil
}

func (p *PostgresBackend) ModulesGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ModulesResponse, error) {
	module, err := modulesSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Provider)
	if err != nil {
		return nil, err
	}
	if module == nil {
		return nil, fmt.Errorf("module not found")
	}

	resp := &models.ModulesResponse{
		Data: models.ModulesDataResponse{
			ID:   module.ID,
			Type: "registry-modules",
			Attributes: models.ModulesAttributesResponse{
				Name:         module.Name,
				Namespace:    module.Namespace,
				RegistryName: module.RegistryName,
				Provider:     module.Provider,
			},
		},
	}

	return resp, nil
}
