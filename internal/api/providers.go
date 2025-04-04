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
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *ProvidersAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}
