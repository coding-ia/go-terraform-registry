package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func errorResponse(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"errors": "Unable to process request",
	})
}

func errorResponseWithMessage(c *gin.Context, status string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"errors": []string{
			status,
		},
	})
}

func errorResponseErrorNotFound(c *gin.Context, status string) {
	c.JSON(http.StatusNotFound, gin.H{
		"errors": []string{
			status,
		},
	})
}
