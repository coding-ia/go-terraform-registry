package selector

import (
	"context"
	backendbase "go-terraform-registry/internal/backend"
	badgerdbbackend "go-terraform-registry/internal/backend/badgerdb_backend"
	dynamodbbackend "go-terraform-registry/internal/backend/dynamodb_backend"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"go-terraform-registry/internal/storage/local_storage"
	"go-terraform-registry/internal/storage/s3_storage"
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

	_ = selected.ConfigureBackend(ctx)

	return selected
}

func SelectStorage(ctx context.Context, config config.RegistryConfig) storage.RegistryProviderStorage {
	var selected storage.RegistryProviderStorage

	switch config.StorageBackend {
	case "s3":
		selected = s3_storage.NewS3Storage(config)
	case "local":
		selected = local_storage.NewLocalStorage(config)
	default:

		selected = s3_storage.NewS3Storage(config)
	}

	_ = selected.ConfigureStorage(ctx)

	return selected
}
