package dynamodb_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
	"net/http"
)

var _ backend.ModuleVersionsBackend = &DynamoDBBackend{}

func (p *DynamoDBBackend) ModuleVersionsCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModuleVersionsRequest) (*models.ModuleVersionsResponse, error) {
	return nil, nil
}

func (p *DynamoDBBackend) ModuleVersionsDelete(ctx context.Context, parameters registrytypes.APIParameters) (int, error) {
	return http.StatusNotFound, nil
}
