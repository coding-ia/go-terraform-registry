package selector

import (
	"context"
	backendbase "go-terraform-registry/internal/backend"
	badgerdbbackend "go-terraform-registry/internal/backend/badgerdb_backend"
	dynamodbbackend "go-terraform-registry/internal/backend/dynamodb_backend"
	"go-terraform-registry/internal/config"
)

func SelectBackend(ctx context.Context, config config.RegistryConfig) backendbase.RegistryProviderBackend {
	var selected backendbase.RegistryProviderBackend

	switch config.Backend {
	case "badgerdb":
		selected = badgerdbbackend.NewBadgerDBBackend(config)
	case "dynamodb":
		selected = dynamodbbackend.NewDynamoDBBackend(config)
	default:
		selected = badgerdbbackend.NewBadgerDBBackend(config)
	}

	selected.ConfigureBackend(ctx)

	return selected
}
