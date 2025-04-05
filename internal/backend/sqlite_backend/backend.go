package sqlite_backend

import (
	"context"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
)

var _ backend.BackendLifecycle = &SQLiteBackend{}

type SQLiteBackend struct {
	Config config.RegistryConfig
}

func NewSQLiteBackend(_ context.Context, config config.RegistryConfig) (*backend.Backend, error) {
	b := &SQLiteBackend{
		Config: config,
	}

	return &backend.Backend{
		BackendLifecycle:        b,
		RegistryBackend:         b,
		ProvidersBackend:        b,
		ProviderVersionsBackend: b,
		GPGKeysBackend:          b,
	}, nil
}

func (s *SQLiteBackend) Configure(ctx context.Context) error {
	return nil
}

func (s *SQLiteBackend) Close(ctx context.Context) error {
	return nil
}
