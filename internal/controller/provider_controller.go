package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"net/http"
)

type ProviderController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.RegistryProviderBackend
	Storage storage.RegistryProviderStorage
}

type RegistryProviderController interface {
	ProviderPackage(*gin.Context)
	Versions(*gin.Context)
}

func NewProviderController(r *gin.Engine, config registryconfig.RegistryConfig, backend backend.RegistryProviderBackend, storage storage.RegistryProviderStorage) RegistryProviderController {
	pc := &ProviderController{
		Config:  config,
		Backend: backend,
		Storage: storage,
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

	userParams := registrytypes.UserParameters{
		Organization: p.Config.Organization,
	}

	provider, err := p.Backend.GetProvider(c.Request.Context(), params, userParams)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if provider == nil {
		errorResponseErrorNotFound(c, "Not Found")
		return
	}

	shaSum := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", params.Name, params.Version)
	shaSumSig := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", params.Name, params.Version)

	key := fmt.Sprintf("%s/%s/%s/%s/%s", userParams.Organization, "private", params.Namespace, params.Name, params.Version)
	
	downloadURL, err := p.Storage.GenerateDownloadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, provider.Filename))
	shaSumURL, err := p.Storage.GenerateDownloadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, shaSum))
	shaSumSigURL, err := p.Storage.GenerateDownloadURL(c.Request.Context(), fmt.Sprintf("%s/%s", key, shaSumSig))

	provider.DownloadUrl = downloadURL
	provider.ShasumsUrl = shaSumURL
	provider.ShasumsSignatureUrl = shaSumSigURL

	c.JSON(http.StatusOK, provider)
}

func (p *ProviderController) Versions(c *gin.Context) {
	params := registrytypes.ProviderVersionParameters{
		Namespace: c.Param("ns"),
		Name:      c.Param("name"),
	}

	userParams := registrytypes.UserParameters{
		Organization: p.Config.Organization,
	}

	provider, err := p.Backend.GetProviderVersions(c.Request.Context(), params, userParams)
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
