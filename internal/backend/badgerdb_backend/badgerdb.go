package badgerdb_backend

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"os"
)

var _ backend.RegistryProviderBackend = &BadgerDBBackend{}

type BadgerDBBackend struct {
	DBPath                   string
	GPGTableName             string
	ProviderTableName        string
	ProviderVersionTableName string
	ModuleTableName          string
}

func NewBadgerDBBackend() backend.RegistryProviderBackend {
	return &BadgerDBBackend{}
}

func (b *BadgerDBBackend) ConfigureBackend(_ context.Context) {
	b.DBPath = "registry_db"
	b.GPGTableName = "gpg"
	b.ProviderTableName = "providers"
	b.ProviderVersionTableName = "provider-version"
	b.ModuleTableName = "modules"

	val, ok := os.LookupEnv("BADGER_DB_PATH")
	if ok {
		b.DBPath = val
	}
}

func (b *BadgerDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters) (*models.TerraformProviderPlatformResponse, error) {
	return nil, nil
}

func (b *BadgerDBBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters) (*models.TerraformAvailableProvider, error) {
	return nil, nil
}

func (b *BadgerDBBackend) GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	return nil, nil
}

func (b *BadgerDBBackend) GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
	return nil, nil
}

func (b *BadgerDBBackend) RegistryProviders(ctx context.Context, request models.RegistryProvidersRequest) (*models.RegistryProvidersResponse, error) {
	return nil, nil
}

func (b *BadgerDBBackend) GPGKey(ctx context.Context, request models.GPGKeyRequest) (*models.GPGKeyResponse, error) {
	return nil, nil
}

func (b *BadgerDBBackend) RegistryProviderVersions(ctx context.Context, request models.RegistryProviderVersionsRequest) (*models.RegistryProviderVersionsResponse, error) {
	return nil, nil
}

func (b *BadgerDBBackend) RegistryProviderVersionPlatforms(ctx context.Context, request models.RegistryProviderVersionPlatformsRequest) (*models.RegistryProviderVersionPlatformsResponse, error) {
	return nil, nil
}

func withBadgerDB(dbPath string, fn func(*badger.DB) error) error {
	opts := badger.DefaultOptions(dbPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	return fn(db)
}
