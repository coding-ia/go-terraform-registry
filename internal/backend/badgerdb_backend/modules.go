package badgerdb_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ModulesBackend = &BadgerDBBackend{}

func (b *BadgerDBBackend) ModulesCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModulesRequest) (*models.ModulesResponse, error) {
	return nil, nil
}

func (b *BadgerDBBackend) ModulesGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ModulesResponse, error) {
	return nil, nil
}
