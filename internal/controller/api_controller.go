package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/models"
	"go-terraform-registry/internal/storage"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
	"strings"
)

type APIController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.RegistryProviderBackend
	Storage storage.RegistryProviderStorage
}

type RegistryAPIController interface {
	RegistryProviders(c *gin.Context)
	RegistryProviderVersions(c *gin.Context)
	RegistryProviderVersionPlatforms(c *gin.Context)
}

func NewAPIController(r *gin.Engine, config registryconfig.RegistryConfig, backend backend.RegistryProviderBackend, storage storage.RegistryProviderStorage) RegistryAPIController {
	ac := &APIController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	endpoint := r.Group("/api")

	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:ns/:name/versions", ac.RegistryProviderVersions)
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:ns/:name/versions/:version/platforms", ac.RegistryProviderVersionPlatforms)

	providersAPI := api.ProvidersAPI{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}
	endpoint.GET("/v2/organizations/:organization_name/registry-providers", providersAPI.List)
	endpoint.POST("/v2/organizations/:organization/registry-providers", providersAPI.Create)
	endpoint.GET("/v2/organizations/:organization_name/registry-providers/:registry_name/:namespace/:name", providersAPI.Get)
	endpoint.DELETE("/v2/organizations/:organization_name/registry-providers/:registry_name/:namespace/:name", providersAPI.Delete)

	gpgKeysAPI := api.GPGKeysAPI{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}
	endpoint.GET("/registry/:registry_name/v2/gpg-keys", gpgKeysAPI.List)
	endpoint.POST("/registry/private/v2/gpg-keys", gpgKeysAPI.Add)
	endpoint.GET("/registry/:registry_name/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Get)
	endpoint.PATCH("/registry/:registry_name/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Update)
	endpoint.DELETE("/registry/:registry_name/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Delete)

	return ac
}

func (a *APIController) RegistryProviders(c *gin.Context) {

}

func (a *APIController) RegistryProviderVersions(c *gin.Context) {
	var req models.RegistryProviderVersionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("ns")
	name := c.Param("name")

	if registry != "private" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "registry must be private"})
		return
	}

	if !strings.EqualFold(organization, namespace) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "namespace must match organization"})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
	}

	resp, err := a.Backend.RegistryProviderVersions(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
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

	c.JSON(http.StatusCreated, resp)
}

func (a *APIController) RegistryProviderVersionPlatforms(c *gin.Context) {
	var req models.RegistryProviderVersionPlatformsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("ns")
	name := c.Param("name")
	version := c.Param("version")

	if registry != "private" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "registry must be private"})
		return
	}

	if !strings.EqualFold(organization, namespace) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "namespace must match organization"})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
		Version:      version,
	}

	resp, err := a.Backend.RegistryProviderVersionPlatforms(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Version)
	uploadURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, req.Data.Attributes.Filename))

	resp.Data.Links = models.RegistryProviderVersionPlatformsLinks{
		ProviderBinaryUpload: uploadURL,
	}

	c.JSON(http.StatusCreated, resp)
}
