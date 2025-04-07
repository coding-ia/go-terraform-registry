package postgres_backend

import (
	"context"
	"fmt"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ProviderVersionsBackend = &PostgresBackend{}

func (p *PostgresBackend) ProviderVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionsRequest) (*models.ProviderVersionsResponse, error) {
	provider, err := providersSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	gpgKey, err := gpgSelect(ctx, p.db, request.Data.Attributes.KeyID, parameters.Namespace)
	if err != nil {
		return nil, err
	}
	if gpgKey == nil {
		return nil, fmt.Errorf("gpg key not found")
	}

	pv := &ProviderVersion{
		ProviderID: provider.ID,
		GPGKeyID:   gpgKey.ID,
		Version:    request.Data.Attributes.Version,
		MetaData: ProviderVersionMetaData{
			Protocols: request.Data.Attributes.Protocols,
		},
	}

	err = providerVersionsInsert(ctx, p.db, pv)
	if err != nil {
		return nil, err
	}

	resp := &models.ProviderVersionsResponse{
		Data: models.ProviderVersionsDataResponse{
			ID:   pv.ID,
			Type: "registry-provider-versions",
			Attributes: models.ProviderVersionsAttributesResponse{
				Version:   pv.Version,
				Protocols: request.Data.Attributes.Protocols,
				KeyID:     pv.GPGKeyID,
			},
		},
	}

	return resp, nil
}

func (p *PostgresBackend) ProviderVersionsGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ProviderVersionsResponse, error) {
	provider, err := providersSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("no provider found for version")
	}

	providerVersion, err := providerVersionSelect(ctx, p.db, provider.ID, parameters.Version)
	if err != nil {
		return nil, err
	}
	if providerVersion == nil {
		return nil, fmt.Errorf("provider version not found")
	}

	resp := &models.ProviderVersionsResponse{
		Data: models.ProviderVersionsDataResponse{
			ID:   providerVersion.ID,
			Type: "registry-provider-versions",
			Attributes: models.ProviderVersionsAttributesResponse{
				Version:   providerVersion.Version,
				Protocols: providerVersion.MetaData.Protocols,
				KeyID:     providerVersion.GPGKeyID,
			},
		},
	}

	return resp, nil
}

func (p *PostgresBackend) ProviderVersionPlatformsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ProviderVersionPlatformsRequest) (*models.ProviderVersionPlatformsResponse, error) {
	provider, err := providersSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	pv, err := providerVersionSelect(ctx, p.db, provider.ID, parameters.Version)
	if err != nil {
		return nil, err
	}

	platform := &ProviderPlatform{
		OS:                request.Data.Attributes.OS,
		Arch:              request.Data.Attributes.Arch,
		SHASum:            request.Data.Attributes.Shasum,
		Filename:          request.Data.Attributes.Filename,
		ProviderVersionID: pv.ID,
	}
	err = providerVersionPlatformInsert(ctx, p.db, platform)
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
