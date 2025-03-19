package config

import (
	"context"
	backendbase "go-terraform-registry/internal/backend"
	badgerdbbackend "go-terraform-registry/internal/backend/badgerdb_backend"
	dynamodbbackend "go-terraform-registry/internal/backend/dynamodb_backend"
)

func SelectBackend(ctx context.Context, backend string) backendbase.RegistryProviderBackend {
	var selected backendbase.RegistryProviderBackend

	switch backend {
	case "badgerdb":
		selected = badgerdbbackend.NewBadgerDBBackend()
	case "dynamodb":
		selected = dynamodbbackend.NewDynamoDBBackend()
	default:
		selected = badgerdbbackend.NewBadgerDBBackend()
	}

	selected.ConfigureBackend(ctx)

	return selected
}
