package postgres_backend

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"os"
)

var _ backend.BackendLifecycle = &PostgresBackend{}

type PostgresBackend struct {
	Config config.RegistryConfig

	db *pgxpool.Pool
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
		ModulesBackend:          b,
		ModuleVersionsBackend:   b,
		GPGKeysBackend:          b,
	}, nil
}

func (p *PostgresBackend) Configure(ctx context.Context) error {
	connectionString := os.Getenv("DATABASE_URL")
	pgxConfig, err := pgxpool.ParseConfig(connectionString)

	// Future use for refreshing credentials (IAM)
	pgxConfig.BeforeConnect = func(ctx context.Context, connCfg *pgx.ConnConfig) error {
		return nil
	}

	db, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	p.db = db

	return nil
}

func (p *PostgresBackend) Close(_ context.Context) error {
	p.db.Close()

	return nil
}
