package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api/models"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
	"strings"
)

type ProviderVersionsAPI api

func (a *ProviderVersionsAPI) CreateVersion(c *gin.Context) {
	var req models.ProviderVersionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("namespace")
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

	resp, err := a.Backend.ProviderVersionsCreate(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	shaSum := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", parameters.Name, req.Data.Attributes.Version)
	shaSumSig := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", parameters.Name, req.Data.Attributes.Version)

	key := fmt.Sprintf("%s/%s/%s/%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, req.Data.Attributes.Version)

	shaSumURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, shaSum))
	shaSumSigURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, shaSumSig))

	resp.Data.Links = models.ProviderVersionsLinksResponse{
		ShasumsUpload:    &shaSumURL,
		ShasumsSigUpload: &shaSumSigURL,
	}

	c.JSON(http.StatusCreated, resp)
}

func (a *ProviderVersionsAPI) ListVersions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProviderVersionsAPI) GetVersion(c *gin.Context) {
	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("namespace")
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

	resp, err := a.Backend.ProviderVersionsGet(c.Request.Context(), parameters)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (a *ProviderVersionsAPI) DeleteVersion(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProviderVersionsAPI) CreatePlatform(c *gin.Context) {
	var req models.ProviderVersionPlatformsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("namespace")
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

	resp, err := a.Backend.ProviderVersionPlatformsCreate(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Version)
	uploadURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, req.Data.Attributes.Filename))

	resp.Data.Links = models.ProviderVersionPlatformsLinksResponse{
		ProviderBinaryUpload: uploadURL,
	}

	c.JSON(http.StatusCreated, resp)
}

func (a *ProviderVersionsAPI) ListPlatform(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProviderVersionsAPI) GetPlatform(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProviderVersionsAPI) DeletePlatform(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}
