package auth

import (
	"github.com/gin-gonic/gin"
	registryconfig "go-terraform-registry/internal/config"
	"net/http"
)

type Authentication struct {
	Config registryconfig.RegistryConfig
}

type AuthenticationMiddleware interface {
	AuthenticationHandler() gin.HandlerFunc
}

func NewAuthenticationMiddleware(config registryconfig.RegistryConfig) AuthenticationMiddleware {
	return &Authentication{
		Config: config,
	}
}

func (a *Authentication) AuthenticationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

	}
}
