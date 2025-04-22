package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/response"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
	"strings"
)

type ModulesAPI api

func (a *ModulesAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req models.ModulesRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	organization := chi.URLParam(r, "organization")

	if !strings.EqualFold(organization, req.Data.Attributes.Namespace) {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: "namespace must match organization",
		})
		return
	}

	parameters := registrytypes.APIParameters{
		Organization: organization,
	}

	resp, err := a.Backend.ModulesCreate(r.Context(), parameters, req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *ModulesAPI) Get(w http.ResponseWriter, r *http.Request) {
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

	resp, err := a.Backend.ModulesGet(r.Context(), parameters)
	if err != nil {
		response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, resp)
}
