package storage

import "context"

type RegistryProviderStorage interface {
	ConfigureStorage(ctx context.Context)
}
