package postgres_backend

import (
	"context"
	"errors"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/pgp"
)

var _ backend.GPGKeysBackend = &PostgresBackend{}

func (p *PostgresBackend) GPGKeysList(ctx context.Context, namespaceFilter string, pageNumber *int, pageSize *int) (*models.GPGKeysListResponse, error) {
	return nil, errors.New("This endpoint is not implemented yet.")
}

func (p *PostgresBackend) GPGKeysAdd(ctx context.Context, request models.GPGKeysRequest) (*models.GPGKeysResponse, error) {
	keyId := pgp.GetKeyID(request.Data.Attributes.AsciiArmor)

	gpg := GPGKey{
		Namespace:  request.Data.Attributes.Namespace,
		KeyID:      keyId[0],
		AsciiArmor: request.Data.Attributes.AsciiArmor,
	}

	err := gpgInsert(ctx, p.db, gpg)
	if err != nil {
		return nil, err
	}

	resp := &models.GPGKeysResponse{
		Data: models.GPGKeysDataResponse{
			ID: keyId[0],
			Attributes: models.GPGKeysAttributesResponse{
				AsciiArmor: request.Data.Attributes.AsciiArmor,
				KeyID:      keyId[0],
				Namespace:  request.Data.Attributes.Namespace,
			},
		},
	}

	return resp, nil
}
