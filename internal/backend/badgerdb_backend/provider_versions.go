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

var _ backend.ProviderVersionsBackend = &BadgerDBBackend{}

func (b *BadgerDBBackend) ProviderVersionsList(ctx context.Context, parameters registrytypes.APIParameters) (*models.ProviderVersionsListResponse, error) {
	return nil, nil
}

func (b *BadgerDBBackend) ProviderVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionsRequest) (*models.ProviderVersionsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.Tables.ProviderTableName, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	var gpg GPGKey
	gpgKey := fmt.Sprintf("%s:%s:%s", b.Tables.GPGTableName, parameters.Namespace, request.Data.Attributes.KeyID)
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return gpgGet(db, gpgKey, &gpg)
	})
	if err != nil {
		return nil, err
	}

	var pv ProviderVersion
	pvKey := fmt.Sprintf("%s:%s:%s", b.Tables.ProviderVersionTableName, p.ID, request.Data.Attributes.Version)
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		err := providerVersionGet(db, pvKey, &pv)
		if err != nil {
			return err
		}

		if pv.ID == "" {
			newUUID := uuid.New()
			pv = ProviderVersion{
				ID:            newUUID.String(),
				Version:       request.Data.Attributes.Version,
				Protocols:     request.Data.Attributes.Protocols,
				GPGKeyID:      request.Data.Attributes.KeyID,
				GPGASCIIArmor: gpg.AsciiArmor,
			}
		}

		return providerVersionSet(db, pvKey, pv)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.ProviderVersionsResponse{
		Data: models.ProviderVersionsDataResponse{
			ID:   pv.ID,
			Type: "registry-provider-versions",
			Attributes: models.ProviderVersionsAttributesResponse{
				Version:   pv.Version,
				Protocols: pv.Protocols,
				KeyID:     pv.GPGKeyID,
			},
		},
	}

	return resp, nil
}

func (b *BadgerDBBackend) ProviderVersionsGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ProviderVersionsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.Tables.ProviderTableName, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	pvKey := fmt.Sprintf("%s:%s:%s", b.Tables.ProviderVersionTableName, p.ID, parameters.Version)
	var pv ProviderVersion
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionGet(db, pvKey, &pv)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.ProviderVersionsResponse{
		Data: models.ProviderVersionsDataResponse{
			ID:   pv.ID,
			Type: "registry-provider-versions",
			Attributes: models.ProviderVersionsAttributesResponse{
				Version:   pv.Version,
				Protocols: pv.Protocols,
				KeyID:     pv.GPGKeyID,
			},
		},
	}

	return resp, nil
}

func (b *BadgerDBBackend) ProviderVersionPlatformsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionPlatformsRequest) (*models.ProviderVersionPlatformsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.Tables.ProviderTableName, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	pvKey := fmt.Sprintf("%s:%s:%s", b.Tables.ProviderVersionTableName, p.ID, parameters.Version)
	var pv ProviderVersion
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionGet(db, pvKey, &pv)
	})
	if err != nil {
		return nil, err
	}

	duplicate := duplicatePlatform(pv.Platform, request.Data.Attributes.OS, request.Data.Attributes.Arch)
	if duplicate {
		return nil, fmt.Errorf("duplicate platform exists matching OS: [%s] -- Architecture [%s]", request.Data.Attributes.OS, request.Data.Attributes.Arch)
	}

	newUUID := uuid.New()
	platform := ProviderPlatform{
		ID:       newUUID.String(),
		OS:       request.Data.Attributes.OS,
		Arch:     request.Data.Attributes.Arch,
		SHASum:   request.Data.Attributes.Shasum,
		Filename: request.Data.Attributes.Filename,
	}

	pv.Platform = append(pv.Platform, platform)
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionSet(db, pvKey, pv)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.ProviderVersionPlatformsResponse{
		Data: models.ProviderVersionPlatformsDataResponse{
			ID:   platform.ID,
			Type: "registry-provider-platforms",
			Attributes: models.ProviderVersionPlatformsAttributesResponse{
				OS:       platform.OS,
				Arch:     platform.Arch,
				Shasum:   platform.SHASum,
				Filename: platform.Filename,
			},
		},
	}

	return resp, nil
}

func (b *BadgerDBBackend) ProviderVersionsDelete(ctx context.Context, parameters registrytypes.APIParameters) (int, error) {
	return -1, nil
}
