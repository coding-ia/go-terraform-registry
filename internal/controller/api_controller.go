package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	"go-terraform-registry/internal/storage"
	registrytypes "go-terraform-registry/internal/types"
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

	organization := c.Param("organization")
	parameters := registrytypes.APIParameters{
		Organization: organization,
	}

	resp, err := a.Backend.RegistryProviders(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (a *APIController) GPGKeys(c *gin.Context) {
	var req models.GPGKeyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.Backend.GPGKey(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (a *APIController) RegistryProviderVersions(c *gin.Context) {
	var req models.RegistryProviderVersionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("ns")
	name := c.Param("name")

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
	}

	resp, err := a.Backend.RegistryProviderVersions(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shaSum := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", parameters.Name, req.Data.Attributes.Version)
	shaSumSig := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", parameters.Name, req.Data.Attributes.Version)

	key := fmt.Sprintf("%s/%s/%s/%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, req.Data.Attributes.Version)

	shaSumURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, shaSum))
	shaSumSigURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, shaSumSig))

	resp.Data.Links = models.RegistryProviderVersionsResponseLinks{
		ShasumsUpload:    shaSumURL,
		ShasumsSigUpload: shaSumSigURL,
	}

	c.JSON(http.StatusOK, resp)
}

func (a *APIController) RegistryProviderVersionPlatforms(c *gin.Context) {
	var req models.RegistryProviderVersionPlatformsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("ns")
	name := c.Param("name")
	version := c.Param("version")

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
		Version:      version,
	}

	resp, err := a.Backend.RegistryProviderVersionPlatforms(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Version)
	uploadURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, req.Data.Attributes.Filename))

	resp.Data.Links = models.RegistryProviderVersionPlatformsLinks{
		ProviderBinaryUpload: uploadURL,
	}

	c.JSON(http.StatusOK, resp)
}
