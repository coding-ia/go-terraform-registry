package postgres_backend

import (
	"context"
	"fmt"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProvidersBackend = &PostgresBackend{}

func (p *PostgresBackend) ProvidersCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProvidersRequest) (*models.ProvidersResponse, error) {
	provider := &Provider{
		Name:         request.Data.Attributes.Name,
		Namespace:    request.Data.Attributes.Namespace,
		Organization: parameters.Organization,
		RegistryName: request.Data.Attributes.RegistryName,
	}

	err := providersInsert(ctx, p.db, provider)
	if err != nil {
		return nil, err
	}

	resp := &models.ProvidersResponse{
		Data: models.ProvidersDataResponse{
			ID:   provider.ID,
			Type: "registry-providers",
			Attributes: models.ProvidersAttributesResponse{
				Name:         provider.Name,
				Namespace:    provider.Namespace,
				RegistryName: provider.RegistryName,
				Permissions: models.ProvidersPermissionsResponse{
					CanDelete: true,
				},
			},
		},
	}

	return resp, nil
}

func (p *PostgresBackend) ProvidersGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ProvidersResponse, error) {
	provider, err := providersSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	resp := &models.ProvidersResponse{
		Data: models.ProvidersDataResponse{
			ID:   provider.ID,
			Type: "registry-providers",
			Attributes: models.ProvidersAttributesResponse{
				Name:         provider.Name,
				Namespace:    provider.Namespace,
				RegistryName: provider.RegistryName,
				Permissions: models.ProvidersPermissionsResponse{
					CanDelete: true,
				},
			},
		},
	}

	return resp, nil
}
