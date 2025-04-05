package postgres_backend

import (
	"context"
	"database/sql"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
)

var _ backend.BackendLifecycle = &PostgresBackend{}

type PostgresBackend struct {
	Config config.RegistryConfig

	client *sql.DB
}

func NewPostgresBackend(_ context.Context, config config.RegistryConfig) (*backend.Backend, error) {
	b := &PostgresBackend{
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

func (p *PostgresBackend) Configure(ctx context.Context) error {
	return nil
}

func (p *PostgresBackend) Close(ctx context.Context) error {
	return nil
}
