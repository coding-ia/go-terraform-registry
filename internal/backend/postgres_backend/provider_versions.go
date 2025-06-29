package postgres_backend

import (
	"context"
	"fmt"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
)

var _ backend.ProviderVersionsBackend = &PostgresBackend{}

func (p *PostgresBackend) ProviderVersionsList(ctx context.Context, parameters registrytypes.APIParameters) (*models.ProviderVersionsListResponse, error) {
	provider, err := providersSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	versions, pagination, err := providerVersionsList(ctx, p.db, provider.ID)
	if err != nil {
		return nil, err
	}

	resp := &models.ProviderVersionsListResponse{
		Meta: models.Meta{
			Pagination: models.PaginationMeta{
				PageSize:    1,
				CurrentPage: 1,
				TotalPages:  1,
				TotalCount:  pagination.TotalCount,
			},
		},
	}

	for _, version := range *versions {
		versionData := models.ProviderVersionsDataResponse{
			ID:   version.ID,
			Type: "registry-provider-versions",
			Attributes: models.ProviderVersionsAttributesResponse{
				Version:   version.Version,
				Protocols: version.MetaData.Protocols,
				KeyID:     version.GPGKeyID,
			},
			Relationships: models.ProviderVersionsRelationshipsResponse{
				RegistryProvider: models.ProviderVersionsRelationshipSingleResponse{
					Data: models.ProviderVersionsRelationshipDataResponse{
						ID:   version.ProviderID,
						Type: "registry-providers",
					},
				},
			},
		}

		for _, platform := range version.Platforms {
			platformData := models.ProviderVersionsRelationshipDataResponse{
				ID:   platform,
				Type: "registry-provider-platforms",
			}
			versionData.Relationships.Platforms.Data = append(versionData.Relationships.Platforms.Data, platformData)
		}

		resp.Data = append(resp.Data, versionData)
	}

	return resp, nil
}

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

func (p *PostgresBackend) ProviderVersionsDelete(ctx context.Context, parameters registrytypes.APIParameters) (int, error) {
	provider, err := providersSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)
	if err != nil {
		return -1, err
	}
	if provider == nil {
		return -1, fmt.Errorf("provider not found")
	}
	err = providerVersionDelete(ctx, p.db, provider.ID, parameters.Version)
	if err != nil {
		return http.StatusNotFound, err
	}
	return http.StatusNoContent, nil
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
