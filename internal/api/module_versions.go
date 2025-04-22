package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/response"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"net/http"
	"strings"
)

type ModuleVersionsAPI api

func (a *ModuleVersionsAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req models.ModuleVersionsRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	organization := chi.URLParam(r, "organization")
	registry := chi.URLParam(r, "registry")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	if !strings.EqualFold(organization, namespace) {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "namespace must match organization",
		})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
		Provider:     provider,
	}

	resp, err := a.Backend.ModuleVersionsCreate(r.Context(), parameters, req)
	if err != nil {
		response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", "modules", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Provider, req.Data.Attributes.Version)
	file := fmt.Sprintf("terraform-%s-%s-%s.tar.gz", parameters.Provider, parameters.Name, req.Data.Attributes.Version)
	moduleURL, err := a.Storage.GenerateUploadURL(r.Context(), fmt.Sprintf("%s/%s", key, file))

	resp.Data.Links = models.ModuleVersionsLinksResponse{
		Upload: moduleURL,
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *ModuleVersionsAPI) Delete(w http.ResponseWriter, r *http.Request) {
	organization := chi.URLParam(r, "organization")
	registry := chi.URLParam(r, "registry")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	parameters := registrytypes.APIParameters{
		Organization: organization,
		Registry:     registry,
		Namespace:    namespace,
		Name:         name,
		Provider:     provider,
		Version:      version,
	}

	statusCode, err := a.Backend.ModuleVersionsDelete(r.Context(), parameters)
	if err != nil {
		response.JsonResponse(w, statusCode, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", "modules", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Provider, parameters.Version)
	file := fmt.Sprintf("terraform-%s-%s-%s.tar.gz", parameters.Provider, parameters.Name, parameters.Version)
	err = a.Storage.RemoveFile(r.Context(), fmt.Sprintf("%s/%s", key, file))
	if err != nil {
		log.Printf("Error removing file: %s", err.Error())
	}

	w.WriteHeader(statusCode)
}
