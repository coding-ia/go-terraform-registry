package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api/models"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
	"strings"
)

type ModuleVersionsAPI api

func (a *ModuleVersionsAPI) Create(c *gin.Context) {
	var req models.ModuleVersionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("namespace")
	name := c.Param("name")
	provider := c.Param("provider")

	if !strings.EqualFold(organization, namespace) {
		c.JSON(http.StatusNotFound, gin.H{"error": "namespace must match organization"})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
		Provider:     provider,
	}

	resp, err := a.Backend.ModuleVersionsCreate(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", "modules", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Provider, req.Data.Attributes.Version)
	file := fmt.Sprintf("terraform-%s-%s-%s.tar.gz", parameters.Provider, parameters.Name, req.Data.Attributes.Version)
	moduleURL, err := a.Storage.GenerateUploadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, file))

	resp.Data.Links = models.ModuleVersionsLinksResponse{
		Upload: moduleURL,
	}

	c.JSON(http.StatusCreated, resp)
}
