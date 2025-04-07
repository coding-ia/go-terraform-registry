package postgres_backend

import (
	"context"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"strings"
)

var _ backend.RegistryBackend = &PostgresBackend{}

func (p *PostgresBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error) {
	release, err := getProviderRelease(ctx, p.db, userParameters.Organization, "private", parameters.Namespace, parameters.Name, parameters.Version)
	if err != nil {
		return nil, err
	}

	response := &models.TerraformProviderPlatformResponse{
		Protocols: release.Protocols,
		SigningKeys: models.SigningKeys{
			GPGPublicKeys: []models.GPGPublicKeys{
				{
					KeyId:      release.GPGKeyID,
					AsciiArmor: release.GPGASCIIArmor,
				},
			},
		},
	}

	for _, platform := range release.Platforms {
		if strings.EqualFold(platform.OS, parameters.OS) &&
			strings.EqualFold(platform.Arch, parameters.Architecture) {
			response.Filename = platform.Filename
			response.Shasum = platform.SHASum
			response.OS = platform.OS
			response.Arch = platform.Arch

			return response, nil
		}
	}

	return nil, nil
}

func (p *PostgresBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters, userParameters registrytypes.UserParameters) (*models.TerraformAvailableProvider, error) {
	releases, err := getProviderReleases(ctx, p.db, userParameters.Organization, "private", parameters.Namespace, parameters.Name)
	if err != nil {
		return nil, err
	}

	var versions []models.TerraformAvailableVersion
	for _, r := range releases {
		v := models.TerraformAvailableVersion{
			Version:   r.Version,
			Protocols: r.Protocols,
		}

		for _, rp := range r.Platforms {
			platform := models.TerraformAvailablePlatform{
				OS:   rp.OS,
				Arch: rp.Arch,
			}
			v.Platforms = append(v.Platforms, platform)
		}
		versions = append(versions, v)
	}

	provider := &models.TerraformAvailableProvider{
		Versions: versions,
	}

	return provider, nil
}

func (p *PostgresBackend) GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	return nil, nil
}

func (p *PostgresBackend) GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
	return nil, nil
}
