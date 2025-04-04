package dynamodb_backend

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProvidersBackend = &DynamoDBBackend{}

func (d *DynamoDBBackend) ProvidersCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProvidersRequest) (*models.ProvidersResponse, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", parameters.Organization, request.Data.Attributes.RegistryName, request.Data.Attributes.Namespace, request.Data.Attributes.Name)

	p, _ := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if p == nil {
		newUUID := uuid.New()
		p = &Provider{
			ID: newUUID.String(),
		}
		err := setProvider(ctx, d.client, d.Tables.ProviderTableName, key, *p)
		if err != nil {
			return nil, err
		}
	}

	resp := &models.ProvidersResponse{
		Data: models.ProvidersDataResponse{
			ID:   p.ID,
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
