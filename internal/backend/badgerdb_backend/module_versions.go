package badgerdb_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
)

var _ backend.ModuleVersionsBackend = &BadgerDBBackend{}

func (p *BadgerDBBackend) ModuleVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModuleVersionsRequest) (*models.ModuleVersionsResponse, error) {
	return nil, nil
}

func (p *BadgerDBBackend) ModuleVersionsDelete(ctx context.Context, parameters registrytypes.APIParameters) (int, error) {
	return http.StatusNotFound, nil
}
