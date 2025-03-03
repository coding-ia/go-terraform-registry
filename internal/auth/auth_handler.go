package auth

import (
	"github.com/gin-gonic/gin"
	registryconfig "go-terraform-registry/internal/config"
	"net/http"
	"strings"
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

		parts := strings.Fields(authHeader)
		token, err := GetJWTToken(parts[1], []byte(a.Config.TokenEncryptionKey))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error parsing token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
