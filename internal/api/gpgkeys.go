package api

import (
	"encoding/json"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/response"
	"net/http"
	"strconv"
)

type GPGKeysAPI api

func (a *GPGKeysAPI) List(w http.ResponseWriter, r *http.Request) {
	queryNamespace := r.URL.Query().Get("filter[namespace]")

	var pageNumber *int
	if r.URL.Query().Has("page[number]") {
		val, _ := strconv.Atoi(r.URL.Query().Get("page[number]"))
		pageNumber = &val
	}

	var pageSize *int
	if r.URL.Query().Has("page[size]") {
		val, _ := strconv.Atoi(r.URL.Query().Get("page[size]"))
		pageSize = &val
	}

	resp, err := a.Backend.GPGKeysList(r.Context(), queryNamespace, pageNumber, pageSize)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, resp)
}

func (a *GPGKeysAPI) Add(w http.ResponseWriter, r *http.Request) {
	var req models.GPGKeysRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	resp, err := a.Backend.GPGKeysAdd(r.Context(), req)
	if err != nil {
		response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response.JsonResponse(w, http.StatusCreated, resp)
}

func (a *GPGKeysAPI) Get(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}

func (a *GPGKeysAPI) Update(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}

func (a *GPGKeysAPI) Delete(w http.ResponseWriter, _ *http.Request) {
	response.JsonResponse(w, http.StatusNotImplemented, response.ErrorResponse{
		Error: "This endpoint is not implemented yet.",
	})
}
