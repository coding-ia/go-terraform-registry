package controller

import (
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"net/http"
)

type ModuleController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.Backend
	Storage storage.RegistryProviderStorage
}

type RegistryModuleController interface {
	ModuleDownload(*gin.Context)
	Versions(*gin.Context)
}

func NewModuleController(r *gin.Engine, config registryconfig.RegistryConfig, backend backend.Backend, storage storage.RegistryProviderStorage) RegistryModuleController {
	mc := &ModuleController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	modules := r.Group("/terraform/modules/v1")

	if !config.AllowAnonymousAccess {
		handler := auth.NewAuthenticationMiddleware(config)
		modules.Use(handler.AuthenticationHandler())
	}

	modules.GET("/:ns/:name/:system/versions", mc.Versions)
	modules.GET("/:ns/:name/:system/:version/download", mc.ModuleDownload)

	return mc
}

func (m *ModuleController) ModuleDownload(c *gin.Context) {
	params := registrytypes.ModuleDownloadParameters{
		Namespace: c.Param("ns"),
		Name:      c.Param("name"),
		System:    c.Param("system"),
		Version:   c.Param("version"),
	}

	path, err := m.Backend.GetModuleDownload(c.Request.Context(), params)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if path == nil {
		errorResponseErrorNotFound(c, "Not Found")
		return
	}

	uri, err := m.Storage.GenerateDownloadURL(c.Request.Context(), *path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating download url."})
	}

	c.Header("X-Terraform-Get", uri)
	c.Status(http.StatusNoContent)
}

func (m *ModuleController) Versions(c *gin.Context) {
	params := registrytypes.ModuleVersionParameters{
		Namespace: c.Param("ns"),
		Name:      c.Param("name"),
		System:    c.Param("system"),
	}

	module, err := m.Backend.GetModuleVersions(c.Request.Context(), params)
	if err != nil {
		log.Printf(err.Error())
		errorResponse(c)
		return
	}

	if module == nil {
		errorResponseErrorNotFound(c, "Not Found")
		return
	}

	c.JSON(http.StatusOK, module)
}
