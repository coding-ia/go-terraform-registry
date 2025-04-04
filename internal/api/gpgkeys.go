package api

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api/models"
	"net/http"
)

type GPGKeysAPI api

func (a *GPGKeysAPI) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *GPGKeysAPI) Add(c *gin.Context) {
	var req models.GPGKeysRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.Backend.GPGKeysAdd(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (a *GPGKeysAPI) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *GPGKeysAPI) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}

func (a *GPGKeysAPI) Delete(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}
