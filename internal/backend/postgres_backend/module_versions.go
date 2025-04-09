package postgres_backend

import (
	"context"
	"fmt"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
)

var _ backend.ModuleVersionsBackend = &PostgresBackend{}

func (p *PostgresBackend) ModuleVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModuleVersionsRequest) (*models.ModuleVersionsResponse, error) {
	module, err := modulesSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Provider)
	if err != nil {
		return nil, err
	}
	if module == nil {
		return nil, fmt.Errorf("module not found")
	}

	mv := &ModuleVersion{
		ModuleID:  module.ID,
		Version:   request.Data.Attributes.Version,
		CommitSHA: request.Data.Attributes.CommitSHA,
	}
	err = moduleVersionsInsert(ctx, p.db, mv)
	if err != nil {
		return nil, err
	}

	resp := &models.ModuleVersionsResponse{
		Data: models.ModuleVersionsDataResponse{
			Type: "registry-module-versions",
			Attributes: models.ModuleVersionsAttributesResponse{
				Version: mv.Version,
			},
		},
	}

	return resp, nil
}

func (p *PostgresBackend) ModuleVersionsDelete(ctx context.Context, parameters registrytypes.APIParameters) (int, error) {
	module, err := modulesSelect(ctx, p.db, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name, parameters.Provider)
	if err != nil {
		return -1, err
	}
	if module == nil {
		return -1, fmt.Errorf("module not found")
	}
	err = moduleVersionsDelete(ctx, p.db, module.ID, parameters.Version)
	if err != nil {
		return http.StatusNotFound, err
	}
	return http.StatusNoContent, nil
}
