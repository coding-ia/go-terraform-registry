package controller

import (
	"fmt"
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

type ProviderController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.Backend
	Storage storage.RegistryProviderStorage
}

type RegistryProviderController interface {
	ProviderPackage(http.ResponseWriter, *http.Request)
	Versions(http.ResponseWriter, *http.Request)
}

func NewProviderController(router chi.Router, config registryconfig.RegistryConfig, backend backend.Backend, storage storage.RegistryProviderStorage) RegistryProviderController {
	pc := &ProviderController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	router.Route("/terraform/providers/v1", func(r chi.Router) {
		if !config.AllowAnonymousAccess {
			handler := auth.NewAuthenticationMiddleware(config)
			r.Use(handler.AuthenticationHandlerMiddleware)
		}

		r.Get("/{ns}/{name}/versions", pc.Versions)
		r.Get("/{ns}/{name}/{version}/download/{os}/{arch}", pc.ProviderPackage)
	})

	return pc
}

func (p *ProviderController) ProviderPackage(w http.ResponseWriter, r *http.Request) {
	params := registrytypes.ProviderPackageParameters{
		Namespace:    chi.URLParam(r, "ns"),
		Name:         chi.URLParam(r, "name"),
		Version:      chi.URLParam(r, "version"),
		OS:           chi.URLParam(r, "os"),
		Architecture: chi.URLParam(r, "arch"),
	}

	userParams := registrytypes.UserParameters{
		Organization: params.Namespace,
	}

	provider, err := p.Backend.GetProvider(r.Context(), params, userParams)
	if err != nil {
		log.Printf(err.Error())
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if provider == nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Not Found",
		})
		return
	}

	shaSum := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", params.Name, params.Version)
	shaSumSig := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", params.Name, params.Version)

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s", "providers", userParams.Organization, "private", params.Namespace, params.Name, params.Version)

	downloadURL, err := p.Storage.GenerateDownloadURL(r.Context(), fmt.Sprintf("%s/%s", key, provider.Filename))
	shaSumURL, err := p.Storage.GenerateDownloadURL(r.Context(), fmt.Sprintf("%s/%s", key, shaSum))
	shaSumSigURL, err := p.Storage.GenerateDownloadURL(r.Context(), fmt.Sprintf("%s/%s", key, shaSumSig))

	provider.DownloadUrl = downloadURL
	provider.ShasumsUrl = shaSumURL
	provider.ShasumsSignatureUrl = shaSumSigURL

	response.JsonResponse(w, http.StatusOK, provider)
}

func (p *ProviderController) Versions(w http.ResponseWriter, r *http.Request) {
	params := registrytypes.ProviderVersionParameters{
		Namespace: chi.URLParam(r, "ns"),
		Name:      chi.URLParam(r, "name"),
	}

	userParams := registrytypes.UserParameters{
		Organization: params.Namespace,
	}

	provider, err := p.Backend.GetProviderVersions(r.Context(), params, userParams)
	if err != nil {
		log.Printf(err.Error())
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if provider == nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Not Found",
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, provider)
}
