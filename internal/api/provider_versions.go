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
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProviderVersionsAPI) DeleteVersion(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProviderVersionsAPI) CreatePlatform(c *gin.Context) {

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
