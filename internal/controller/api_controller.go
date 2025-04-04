package controller

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
)

type APIController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.RegistryProviderBackend
	Storage storage.RegistryProviderStorage
}

type RegistryAPIController interface {
}

func NewAPIController(r *gin.Engine, config registryconfig.RegistryConfig, backend backend.RegistryProviderBackend, storage storage.RegistryProviderStorage) RegistryAPIController {
	ac := &APIController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	endpoint := r.Group("/api")

	providerVersionsAPI := api.ProviderVersionsAPI{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions", providerVersionsAPI.CreateVersion)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/", providerVersionsAPI.ListVersions)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version", providerVersionsAPI.GetVersion)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version", providerVersionsAPI.DeleteVersion)
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms", providerVersionsAPI.CreatePlatform)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms", providerVersionsAPI.ListPlatform)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms/:os/:arch", providerVersionsAPI.GetPlatform)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms/:os/:arch", providerVersionsAPI.DeletePlatform)

	providersAPI := api.ProvidersAPI{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}
	endpoint.GET("/v2/organizations/:organization/registry-providers", providersAPI.List)
	endpoint.POST("/v2/organizations/:organization/registry-providers", providersAPI.Create)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name", providersAPI.Get)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name", providersAPI.Delete)

	gpgKeysAPI := api.GPGKeysAPI{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}
	endpoint.GET("/registry/:registry/v2/gpg-keys", gpgKeysAPI.List)
	endpoint.POST("/registry/private/v2/gpg-keys", gpgKeysAPI.Add)
	endpoint.GET("/registry/:registry/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Get)
	endpoint.PATCH("/registry/:registry/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Update)
	endpoint.DELETE("/registry/:registry/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Delete)

	return ac
}
