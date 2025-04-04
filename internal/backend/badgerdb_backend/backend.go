package badgerdb_backend

import (
	"context"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"log"
	"os"
)

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

	configureBackend(b)

	return &backend.Backend{
		RegistryBackend:         b,
		ProvidersBackend:        b,
		ProviderVersionsBackend: b,
		GPGKeysBackend:          b,
	}, nil
}

func configureBackend(badgerDBBackend *BadgerDBBackend) {
	badgerDBBackend.DBPath = "registry_db"
	badgerDBBackend.Tables.GPGTableName = "gpg"
	badgerDBBackend.Tables.ProviderTableName = "providers"
	badgerDBBackend.Tables.ProviderVersionTableName = "provider-version"
	badgerDBBackend.Tables.ModuleTableName = "modules"

	val, ok := os.LookupEnv("BADGER_DB_PATH")
	if ok {
		badgerDBBackend.DBPath = val
	}

	log.Println("Using BadgerDB for backend.")
}
