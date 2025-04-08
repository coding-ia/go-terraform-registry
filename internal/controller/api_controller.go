package controller

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"net/http"
)

type APIController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.Backend
	Storage storage.RegistryProviderStorage
}

type RegistryAPIController interface {
	CreateEndpoints(r *gin.Engine)
	AuthenticateRequest(c *gin.Context)
}

func NewAPIController(config registryconfig.RegistryConfig, backend backend.Backend, storage storage.RegistryProviderStorage) RegistryAPIController {
	ac := &APIController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	return ac
}

func (a *APIController) CreateEndpoints(r *gin.Engine) {
	endpoint := r.Group("/api", a.AuthenticateRequest)

	providerVersionsAPI := api.ProviderVersionsAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions", providerVersionsAPI.CreateVersion)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/", providerVersionsAPI.ListVersions)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version", providerVersionsAPI.GetVersion)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version", providerVersionsAPI.DeleteVersion)
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms", providerVersionsAPI.CreatePlatform)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms", providerVersionsAPI.ListPlatform)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms/:os/:arch", providerVersionsAPI.GetPlatform)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms/:os/:arch", providerVersionsAPI.DeletePlatform)

	providersAPI := api.ProvidersAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.GET("/v2/organizations/:organization/registry-providers", providersAPI.List)
	endpoint.POST("/v2/organizations/:organization/registry-providers", providersAPI.Create)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name", providersAPI.Get)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name", providersAPI.Delete)

	modulesAPI := api.ModulesAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.POST("/v2/organizations/:organization/registry-modules", modulesAPI.Create)
	endpoint.GET("/v2/organizations/:organization/registry-modules/:registry/:namespace/:name/:provider", modulesAPI.Get)

	gpgKeysAPI := api.GPGKeysAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.GET("/registry/:registry/v2/gpg-keys", gpgKeysAPI.List)
	endpoint.POST("/registry/private/v2/gpg-keys", gpgKeysAPI.Add)
	endpoint.GET("/registry/:registry/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Get)
	endpoint.PATCH("/registry/:registry/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Update)
	endpoint.DELETE("/registry/:registry/v2/gpg-keys/:namespace/:key_id", gpgKeysAPI.Delete)
}

func (a *APIController) AuthenticateRequest(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}

	tokenString := authHeader[len(prefix):]
	_, err := auth.GetJWTToken(tokenString, []byte(a.Config.TokenEncryptionKey))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}
}
