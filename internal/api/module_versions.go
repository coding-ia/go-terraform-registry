package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ModuleVersionsAPI api

func (a *ModuleVersionsAPI) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not implemented yet."})
}
