package dynamodb_backend

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProviderVersionsBackend = &DynamoDBBackend{}

func (d *DynamoDBBackend) ProviderVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionsRequest) (*models.ProviderVersionsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	provider, err := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}

	gpg, err := getGPG(ctx, d.client, d.Tables.GPGTableName, parameters.Namespace, request.Data.Attributes.KeyID)
	if err != nil {
		return nil, err
	}
	if gpg == nil {
		return nil, fmt.Errorf("no GPG key found for %s", request.Data.Attributes.KeyID)
	}

	newUUID := uuid.New()
	pv := ProviderVersion{
		ID:            newUUID.String(),
		Version:       request.Data.Attributes.Version,
		Protocols:     request.Data.Attributes.Protocols,
		GPGKeyID:      gpg.KeyID,
		GPGASCIIArmor: gpg.AsciiArmor,
	}
	err = setProviderVersion(ctx, d.client, d.Tables.ProviderVersionTableName, *provider, pv)
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

func (d *DynamoDBBackend) ProviderVersionsGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ProviderVersionsResponse, error) {
	return nil, nil
}

func (d *DynamoDBBackend) ProviderVersionPlatformsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionPlatformsRequest) (*models.ProviderVersionPlatformsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	provider, err := getProvider(ctx, d.client, d.Tables.ProviderTableName, key)
	if err != nil {
		return nil, err
	}

	pv, err := getProviderVersion(ctx, d.client, d.Tables.ProviderVersionTableName, *provider, parameters.Version)
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

	err = appendPlatform(ctx, d.client, d.Tables.ProviderVersionTableName, provider.Provider, parameters.Version, platform)
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
