package api

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api/models"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
	"strings"
)

type ProvidersAPI api

func (a *ProvidersAPI) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProvidersAPI) Create(c *gin.Context) {
	var req models.ProvidersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	organization := c.Param("organization")

	if !strings.EqualFold(organization, req.Data.Attributes.Namespace) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "namespace must match organization"})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
	}

	resp, err := a.Backend.ProvidersCreate(c.Request.Context(), parameters, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (a *ProvidersAPI) Get(c *gin.Context) {
	organization := c.Param("organization")
	registry := c.Param("registry")
	namespace := c.Param("namespace")
	name := c.Param("name")

	if !strings.EqualFold(organization, namespace) {
		c.JSON(http.StatusNotFound, gin.H{"error": "namespace must match organization"})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
	}

	resp, err := a.Backend.ProvidersGet(c.Request.Context(), parameters)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (a *ProvidersAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}
