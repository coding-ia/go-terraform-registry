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

type ProvidersAPI api

func (a *ProvidersAPI) List(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}

func (a *ProvidersAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req models.ProvidersRequest

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

	resp, err := a.Backend.ProvidersCreate(r.Context(), parameters, req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *ProvidersAPI) Get(w http.ResponseWriter, r *http.Request) {
	organization := chi.URLParam(r, "organization")
	registry := chi.URLParam(r, "registry")
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

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

	resp, err := a.Backend.ProvidersGet(r.Context(), parameters)
	if err != nil {
		response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, resp)
}

func (a *ProvidersAPI) Delete(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}
