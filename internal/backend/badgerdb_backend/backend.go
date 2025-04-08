package badgerdb_backend

import (
	"context"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"log"
	"os"
)

var _ backend.BackendLifecycle = &BadgerDBBackend{}

type BadgerDBBackend struct {
	Config config.RegistryConfig
	DBPath string
	Tables BadgerTables
}

type BadgerTables struct {
	GPGTableName             string
	ProviderTableName        string
	ProviderVersionTableName string
	ModuleTableName          string
}

func NewBadgerDBBackend(_ context.Context, config config.RegistryConfig) (*backend.Backend, error) {
	b := &BadgerDBBackend{
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

func (b *BadgerDBBackend) Configure(ctx context.Context) error {
	b.DBPath = "registry_db"
	b.Tables.GPGTableName = "gpg"
	b.Tables.ProviderTableName = "providers"
	b.Tables.ProviderVersionTableName = "provider-version"
	b.Tables.ModuleTableName = "modules"

	val, ok := os.LookupEnv("BADGER_DB_PATH")
	if ok {
		b.DBPath = val
	}

	log.Println("Using BadgerDB for backend.")

	return nil
}

func (b *BadgerDBBackend) Close(ctx context.Context) error {
	return nil
}
