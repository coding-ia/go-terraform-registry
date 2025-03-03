package controller

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"net/http"
)

type ProviderController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.RegistryProviderBackend
}

type RegistryProviderController interface {
	ProviderPackage(*gin.Context)
	Versions(*gin.Context)
}

func NewProviderController(r *gin.Engine, config registryconfig.RegistryConfig, backend backend.RegistryProviderBackend) RegistryProviderController {
	pc := &ProviderController{
		Config:  config,
		Backend: backend,
	}

	providers := r.Group("/terraform/providers/v1")

	if !config.AllowAnonymousAccess {
		handler := auth.NewAuthenticationMiddleware(config)
		providers.Use(handler.AuthenticationHandler())
	}

	providers.GET("/:ns/:name/versions", pc.Versions)
	providers.GET("/:ns/:name/:version/download/:os/:arch", pc.ProviderPackage)

	return pc
}

func (p *ProviderController) ProviderPackage(c *gin.Context) {
	params := registrytypes.ProviderPackageParameters{
		Namespace:    c.Param("ns"),
		Name:         c.Param("name"),
		Version:      c.Param("version"),
		OS:           c.Param("os"),
		Architecture: c.Param("arch"),
	}

	provider, err := p.Backend.GetProvider(c.Request.Context(), params)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if provider == nil {
		errorResponseErrorNotFound(c, "Not Found")
		return
	}

	c.JSON(http.StatusOK, provider)
}

func (p *ProviderController) Versions(c *gin.Context) {
	params := registrytypes.ProviderVersionParameters{
		Namespace: c.Param("ns"),
		Name:      c.Param("name"),
	}

	provider, err := p.Backend.GetProviderVersions(c.Request.Context(), params)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if provider == nil {
		errorResponseErrorNotFound(c, "Not Found")
		return
	}

	c.JSON(http.StatusOK, provider)
}
