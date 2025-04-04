package badgerdb_backend

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProvidersBackend = &BadgerDBBackend{}

func (b *BadgerDBBackend) ProvidersCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProvidersRequest) (*models.ProvidersResponse, error) {
	var p Provider
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.Tables.ProviderTableName, parameters.Organization, request.Data.Attributes.RegistryName, request.Data.Attributes.Namespace, request.Data.Attributes.Name)
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		err := providerGet(db, key, &p)
		if err != nil {
			return err
		}

		if p.ID == "" {
			newUUID := uuid.New()
			p.ID = newUUID.String()
		}

		return providerSet(db, key, p)
	})
	if err != nil {
		return nil, err
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
