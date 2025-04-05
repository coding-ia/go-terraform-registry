package selector

import (
	"context"
	backendbase "go-terraform-registry/internal/backend"
	badgerdbbackend "go-terraform-registry/internal/backend/badgerdb_backend"
	dynamodbbackend "go-terraform-registry/internal/backend/dynamodb_backend"
	sqlitebackend "go-terraform-registry/internal/backend/sqlite_backend"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/storage"
	"go-terraform-registry/internal/storage/local_storage"
	"go-terraform-registry/internal/storage/s3_storage"
	"log"
)

func SelectBackend(ctx context.Context, config config.RegistryConfig) *backendbase.Backend {
	var selected *backendbase.Backend
	var err error

	switch config.Backend {
	case "badgerdb":
		selected, err = badgerdbbackend.NewBadgerDBBackend(ctx, config)
	case "dynamodb":
		selected, err = dynamodbbackend.NewDynamoDBBackend(ctx, config)
	case "sqlite":
		selected, err = sqlitebackend.NewSQLiteBackend(ctx, config)
	default:
		selected, err = badgerdbbackend.NewBadgerDBBackend(ctx, config)
	}
	if err != nil {
		log.Fatal(err)
	}

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
		selected = local_storage.NewLocalStorage(config)
	}

	_ = selected.ConfigureStorage(ctx)

	return selected
}
