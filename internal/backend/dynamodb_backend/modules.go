package dynamodb_backend

import (
	"context"
	"go-terraform-registry/internal/api/models"
	"go-terraform-registry/internal/backend"
	registrytypes "go-terraform-registry/internal/types"
)

var _ backend.ModulesBackend = &DynamoDBBackend{}

func (d *DynamoDBBackend) ModulesCreate(ctx context.Context, parameters registrytypes.APIParameters, request models.ModulesRequest) (*models.ModulesResponse, error) {
	return nil, nil
}

func (d *DynamoDBBackend) ModulesGet(ctx context.Context, parameters registrytypes.APIParameters) (*models.ModulesResponse, error) {
	return nil, nil
}
