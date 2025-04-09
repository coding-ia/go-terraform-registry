package controller

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/api"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"net/http"
	"strings"
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
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions", validateOrganization, providerVersionsAPI.CreateVersion)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/", validateOrganization, providerVersionsAPI.ListVersions)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version", validateOrganization, providerVersionsAPI.GetVersion)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version", validateOrganization, providerVersionsAPI.DeleteVersion)
	endpoint.POST("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms", validateOrganization, providerVersionsAPI.CreatePlatform)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms", validateOrganization, providerVersionsAPI.ListPlatform)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms/:os/:arch", validateOrganization, providerVersionsAPI.GetPlatform)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name/versions/:version/platforms/:os/:arch", validateOrganization, providerVersionsAPI.DeletePlatform)

	providersAPI := api.ProvidersAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.GET("/v2/organizations/:organization/registry-providers", validateOrganization, providersAPI.List)
	endpoint.POST("/v2/organizations/:organization/registry-providers", validateOrganization, providersAPI.Create)
	endpoint.GET("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name", validateOrganization, providersAPI.Get)
	endpoint.DELETE("/v2/organizations/:organization/registry-providers/:registry/:namespace/:name", validateOrganization, providersAPI.Delete)

	modulesAPI := api.ModulesAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.POST("/v2/organizations/:organization/registry-modules", validateOrganization, modulesAPI.Create)
	endpoint.GET("/v2/organizations/:organization/registry-modules/:registry/:namespace/:name/:provider", validateOrganization, modulesAPI.Get)

	moduleVersionsAPI := api.ModuleVersionsAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.POST("/v2/organizations/:organization/registry-modules/:registry/:namespace/:name/:provider/versions", validateOrganization, moduleVersionsAPI.Create)

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
	token, err := auth.GetJWTClaimsToken(tokenString, []byte(a.Config.TokenEncryptionKey))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}

	if token != nil && token.Valid {
		if claims, ok := token.Claims.(*auth.RegistryClaims); ok {
			c.Set("organization", claims.Organization)
		}
	}

	c.Next()
}

func validateOrganization(c *gin.Context) {
	organizationParam := c.Param("organization")
	organization, exist := c.Get("organization")
	if !exist {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "Missing organization"})
		return
	}
	if !strings.EqualFold(organization.(string), organizationParam) {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid token for organization"})
		return
	}
}
