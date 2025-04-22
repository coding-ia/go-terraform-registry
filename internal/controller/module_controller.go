package controller

import (
	"github.com/go-chi/chi/v5"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/response"
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
	ModuleDownload(http.ResponseWriter, *http.Request)
	Versions(http.ResponseWriter, *http.Request)
}

func NewModuleController(router chi.Router, config registryconfig.RegistryConfig, backend backend.Backend, storage storage.RegistryProviderStorage) RegistryModuleController {
	mc := &ModuleController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	router.Route("/terraform/modules/v1", func(r chi.Router) {
		if !config.AllowAnonymousAccess {
			handler := auth.NewAuthenticationMiddleware(config)
			r.Use(handler.AuthenticationHandlerMiddleware)
		}

		r.Get("/{ns}/{name}/{system}/versions", mc.Versions)
		r.Get("/{ns}/{name}/{system}/{version}/download", mc.ModuleDownload)
	})

	return mc
}

func (m *ModuleController) ModuleDownload(w http.ResponseWriter, r *http.Request) {
	params := registrytypes.ModuleDownloadParameters{
		Namespace: chi.URLParam(r, "ns"),
		Name:      chi.URLParam(r, "name"),
		System:    chi.URLParam(r, "system"),
		Version:   chi.URLParam(r, "version"),
	}

	path, err := m.Backend.GetModuleDownload(r.Context(), params)
	if err != nil {
		log.Printf(err.Error())
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if path == nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Not Found",
		})
		return
	}

	uri, err := m.Storage.GenerateDownloadURL(r.Context(), *path)
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Error generating download url.",
		})
	}

	w.Header().Set("X-Terraform-Get", uri)
	w.WriteHeader(http.StatusNoContent)
}

func (m *ModuleController) Versions(w http.ResponseWriter, r *http.Request) {
	params := registrytypes.ModuleVersionParameters{
		Namespace: chi.URLParam(r, "ns"),
		Name:      chi.URLParam(r, "name"),
		System:    chi.URLParam(r, "system"),
	}

	module, err := m.Backend.GetModuleVersions(r.Context(), params)
	if err != nil {
		log.Printf(err.Error())
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if module == nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Not Found",
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, module)
}
