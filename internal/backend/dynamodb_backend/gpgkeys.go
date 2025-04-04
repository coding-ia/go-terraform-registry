package dynamodb_backend

import (
	"context"
	"github.com/google/uuid"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/pgp"
)

var _ backend.GPGKeysBackend = &DynamoDBBackend{}

func (d *DynamoDBBackend) GPGKeysAdd(ctx context.Context, request models.GPGKeysRequest) (*models.GPGKeysResponse, error) {
	newUUID := uuid.New()
	keyId := pgp.GetKeyID(request.Data.Attributes.AsciiArmor)

	gpg := GPGKey{
		Namespace:  request.Data.Attributes.Namespace,
		KeyID:      keyId[0],
		ID:         newUUID.String(),
		AsciiArmor: request.Data.Attributes.AsciiArmor,
	}
	err := setGPG(ctx, d.client, d.Tables.GPGTableName, gpg)
	if err != nil {
		return nil, err
	}

	resp := &models.GPGKeysResponse{
		Data: models.GPGKeysDataResponse{
			ID: newUUID.String(),
			Attributes: models.GPGKeysAttributesResponse{
				AsciiArmor: request.Data.Attributes.AsciiArmor,
				KeyID:      keyId[0],
				Namespace:  request.Data.Attributes.Namespace,
			},
		},
	}

	return resp, nil
}
