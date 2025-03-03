package config

import (
	"context"
	backendbase "go-terraform-registry/internal/backend"
	dynamodbbackend "go-terraform-registry/internal/backend/dynamodb_backend"
)

func SelectBackend(ctx context.Context, backend string) backendbase.RegistryProviderBackend {
	var selected backendbase.RegistryProviderBackend

	switch backend {
	case "dynamodb":
		selected = dynamodbbackend.NewDynamoDBBackend()
	default:
		selected = dynamodbbackend.NewDynamoDBBackend()
	}

	selected.ConfigureBackend(ctx)

	return selected
}
