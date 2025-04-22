package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
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
	Chi     *chi.Mux
}

type RegistryAPIController interface {
	CreateEndpoints(r *gin.Engine, cr *chi.Mux)
	AuthenticateRequest(c *gin.Context)
}

func NewAPIController(config registryconfig.RegistryConfig, backend backend.Backend, storage storage.RegistryProviderStorage, cr *chi.Mux) RegistryAPIController {
	ac := &APIController{
		Config:  config,
		Backend: backend,
		Storage: storage,
		Chi:     cr,
	}

	return ac
}

func (a *APIController) CreateEndpoints(r *gin.Engine, cr *chi.Mux) {
	endpoint := r.Group("/api", a.CHIMigrate, a.AuthenticateRequest)

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
	endpoint.DELETE("/v2/organizations/:organization/registry-modules/:registry/:namespace/:name/:provider/:version", validateOrganization, moduleVersionsAPI.Delete)

	gpgKeysAPI := api.GPGKeysAPI{
		Config:  a.Config,
		Backend: a.Backend,
		Storage: a.Storage,
	}
	endpoint.GET("/registry/:registry/v2/gpg-keys")
	endpoint.POST("/registry/private/v2/gpg-keys")
	endpoint.GET("/registry/:registry/v2/gpg-keys/:namespace/:key_id")
	endpoint.PATCH("/registry/:registry/v2/gpg-keys/:namespace/:key_id")
	endpoint.DELETE("/registry/:registry/v2/gpg-keys/:namespace/:key_id")

	cr.Route("/api", func(r chi.Router) {
		r.Get("/registry/{registry}/v2/gpg-keys", gpgKeysAPI.List)
		r.Post("/registry/{registry}/v2/gpg-keys", gpgKeysAPI.Add)
		r.Get("/registry/{registry}/v2/gpg-keys/{namespace}/{key_id}", gpgKeysAPI.Get)
		r.Patch("/registry/{registry}/v2/gpg-keys/{namespace}/{key_id}", gpgKeysAPI.Update)
		r.Delete("/registry/{registry}/v2/gpg-keys/{namespace}/{key_id}", gpgKeysAPI.Delete)
	})
}

func (a *APIController) CHIMigrate(c *gin.Context) {
	ctx := chi.NewRouteContext()
	match := a.Chi.Match(ctx, c.Request.Method, c.Request.URL.Path)

	if match {
		a.Chi.ServeHTTP(c.Writer, c.Request)
		return
	}

	c.Next()
}

func (a *APIController) AuthenticateRequest(c *gin.Context) {
	ctx := chi.NewRouteContext()
	match := a.Chi.Match(ctx, c.Request.Method, c.Request.URL.Path)

	if match {
		return
	}

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
	c.Next()
}

func isChiRoute(path string) bool {
	if strings.Contains(path, "gpg-keys") {
		return true
	}

	return false
}
