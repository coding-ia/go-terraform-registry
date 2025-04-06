package postgres_backend

import (
	"context"
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

	if provider.ID == "" {
		existingProvider, err := providersSelect(ctx, p.db, parameters.Organization, request.Data.Attributes.RegistryName, request.Data.Attributes.Namespace, request.Data.Attributes.Name)
		if err != nil {
			return nil, err
		}
		provider.ID = existingProvider.ID
	}

	resp := &models.ProvidersResponse{
		Data: models.ProvidersDataResponse{
			ID:   provider.ID,
			Type: "registry-providers",
			Attributes: models.ProvidersAttributesResponse{
				Name:         request.Data.Attributes.Name,
				Namespace:    request.Data.Attributes.Namespace,
				RegistryName: request.Data.Attributes.RegistryName,
				Permissions: models.ProvidersPermissionsResponse{
					CanDelete: true,
				},
			},
		},
	}

	return resp, nil
}
