package controller

import (
	"github.com/gin-gonic/gin"
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
	RegistryProviders(c *gin.Context)
	GPGKeys(c *gin.Context)
	RegistryProviderVersions(c *gin.Context)
	RegistryProviderVersionPlatforms(c *gin.Context)
}

func NewAPIController(r *gin.Engine, config registryconfig.RegistryConfig, backend backend.RegistryProviderBackend, storage storage.RegistryProviderStorage) RegistryAPIController {
	ac := &APIController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	api := r.Group("/api")

	api.POST("/v2/organizations/:organization/registry-providers", ac.RegistryProviders)
	api.POST("/registry/private/v2/gpg-keys", ac.GPGKeys)
	api.POST("/v2/organizations/:organization/registry-providers/:registry/:ns/:name/versions", ac.RegistryProviderVersions)
	api.POST("/v2/organizations/:organization/registry-providers/:registry/:ns/:name/versions/:version/platforms", ac.RegistryProviderVersionPlatforms)

	return ac
}

func (a *APIController) RegistryProviders(c *gin.Context) {

}

func (a *APIController) GPGKeys(c *gin.Context) {

}

func (a *APIController) RegistryProviderVersions(c *gin.Context) {

}

func (a *APIController) RegistryProviderVersionPlatforms(c *gin.Context) {

}
