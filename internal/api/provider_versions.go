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

type ProviderVersionsAPI api

func (a *ProviderVersionsAPI) CreateVersion(w http.ResponseWriter, r *http.Request) {
	var req models.ProviderVersionsRequest

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

	if registry != "private" {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "registry must be private",
		})
		return
	}

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
	}

	resp, err := a.Backend.ProviderVersionsCreate(r.Context(), parameters, req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	shaSum := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS", parameters.Name, req.Data.Attributes.Version)
	shaSumSig := fmt.Sprintf("terraform-provider-%s_%s_SHA256SUMS.sig", parameters.Name, req.Data.Attributes.Version)

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s", "providers", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, req.Data.Attributes.Version)

	shaSumURL, err := a.Storage.GenerateUploadURL(r.Context(), fmt.Sprintf("%s/%s", key, shaSum))
	shaSumSigURL, err := a.Storage.GenerateUploadURL(r.Context(), fmt.Sprintf("%s/%s", key, shaSumSig))

	resp.Data.Links = models.ProviderVersionsLinksResponse{
		ShasumsUpload:    &shaSumURL,
		ShasumsSigUpload: &shaSumSigURL,
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *ProviderVersionsAPI) ListVersions(w http.ResponseWriter, r *http.Request) {
	organization := chi.URLParam(r, "organization")
	registry := chi.URLParam(r, "registry")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	if registry != "private" {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "registry must be private",
		})
		return
	}

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
	}

	resp, err := a.Backend.ProviderVersionsList(r.Context(), parameters)
	if err != nil {
		response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *ProviderVersionsAPI) GetVersion(w http.ResponseWriter, r *http.Request) {
	organization := chi.URLParam(r, "organization")
	registry := chi.URLParam(r, "registry")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")

	if registry != "private" {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "registry must be private",
		})
		return
	}

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
		Version:      version,
	}

	resp, err := a.Backend.ProviderVersionsGet(r.Context(), parameters)
	if err != nil {
		response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, resp)
}

func (a *ProviderVersionsAPI) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	organization := chi.URLParam(r, "organization")
	registry := chi.URLParam(r, "registry")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")

	if registry != "private" {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "registry must be private",
		})
		return
	}

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
		Version:      version,
	}

	statusCode, err := a.Backend.ProviderVersionsDelete(r.Context(), parameters)
	if err != nil {
		response.JsonResponse(w, statusCode, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s", "providers", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Version)
	err = a.Storage.RemoveDirectory(r.Context(), key)
	if err != nil {
		log.Printf("Error removing directory: %s", err.Error())
	}

	w.WriteHeader(statusCode)
}

func (a *ProviderVersionsAPI) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	var req models.ProviderVersionPlatformsRequest

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
	version := chi.URLParam(r, "version")

	if registry != "private" {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "registry must be private",
		})
		return
	}

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
		Version:      version,
	}

	resp, err := a.Backend.ProviderVersionPlatformsCreate(r.Context(), parameters, req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	key := fmt.Sprintf("%s/%s/%s/%s/%s/%s", "providers", parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Version)
	uploadURL, err := a.Storage.GenerateUploadURL(r.Context(), fmt.Sprintf("%s/%s", key, req.Data.Attributes.Filename))

	resp.Data.Links = models.ProviderVersionPlatformsLinksResponse{
		ProviderBinaryUpload: uploadURL,
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *ProviderVersionsAPI) ListPlatform(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}

func (a *ProviderVersionsAPI) GetPlatform(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}

func (a *ProviderVersionsAPI) DeletePlatform(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}
