package controller

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	"go-terraform-registry/internal/storage"
	"net/http"
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
	var req models.RegistryProvidersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response models.RegistryProvidersResponse
	c.JSON(http.StatusOK, response)
}

func (a *APIController) GPGKeys(c *gin.Context) {
	var req models.GPGKeyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response models.GPGKeyResponse
	c.JSON(http.StatusOK, response)
}

func (a *APIController) RegistryProviderVersions(c *gin.Context) {
	var req models.RegistryProviderVersionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response models.RegistryProviderVersionResponse
	c.JSON(http.StatusOK, response)
}

func (a *APIController) RegistryProviderVersionPlatforms(c *gin.Context) {
	var req models.RegistryProviderVersionPlatformRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var response models.RegistryProviderVersionPlatformResponse
	c.JSON(http.StatusOK, response)
}
