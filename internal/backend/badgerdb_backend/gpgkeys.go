package badgerdb_backend

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/pgp"
)

var _ backend.GPGKeysBackend = &BadgerDBBackend{}

func (b *BadgerDBBackend) GPGKeysAdd(ctx context.Context, request models.GPGKeysRequest) (*models.GPGKeysResponse, error) {
	newUUID := uuid.New()
	keyId := pgp.GetKeyID(request.Data.Attributes.AsciiArmor)

	gpg := GPGKey{
		Namespace:  request.Data.Attributes.Namespace,
		KeyID:      keyId[0],
		ID:         newUUID.String(),
		AsciiArmor: request.Data.Attributes.AsciiArmor,
	}

	key := fmt.Sprintf("%s:%s:%s", b.Tables.GPGTableName, request.Data.Attributes.Namespace, keyId[0])
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return gpgSet(db, key, gpg)
	})
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
