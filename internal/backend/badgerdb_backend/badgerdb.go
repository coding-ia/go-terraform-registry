package badgerdb_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"os"
	"strings"
)

var _ backend.RegistryProviderBackend = &BadgerDBBackend{}

type BadgerDBBackend struct {
	DBPath            string
	ProviderTableName string
	ModuleTableName   string
}

func NewBadgerDBBackend() backend.RegistryProviderBackend {
	return &BadgerDBBackend{}
}

func (b *BadgerDBBackend) ConfigureBackend(_ context.Context) {
	b.DBPath = "registry_db"
	b.ProviderTableName = "providers"
	b.ModuleTableName = "modules"

	val, ok := os.LookupEnv("BADGER_DB_PATH")
	if ok {
		b.DBPath = val
	}
}

func (b *BadgerDBBackend) GetProvider(_ context.Context, parameters registrytypes.ProviderPackageParameters) (*models.TerraformProviderPlatformResponse, error) {
	opts := badger.DefaultOptions(b.DBPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer func(client *badger.DB) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	key := fmt.Sprintf("%s:%s/%s:%s", b.ProviderTableName, parameters.Namespace, parameters.Name, parameters.Version)

	var providerVersion ProviderVersion
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &providerVersion)
		})
	})
	if err != nil {
		return nil, err
	}

	response := &models.TerraformProviderPlatformResponse{
		Protocols:           providerVersion.Protocols,
		ShasumsUrl:          providerVersion.SHASUMUrl,
		ShasumsSignatureUrl: providerVersion.SHASUMSigUrl,
		SigningKeys: models.SigningKeys{
			GPGPublicKeys: []models.GPGPublicKeys{
				{
					KeyId:      providerVersion.GPGFingerprint,
					AsciiArmor: providerVersion.GPGASCIIArmor,
				},
			},
		},
	}

	for _, p := range providerVersion.Provider {
		if strings.EqualFold(p.OS, parameters.OS) &&
			strings.EqualFold(p.Architecture, parameters.Architecture) {
			response.Filename = p.Filename
			response.DownloadUrl = p.DownloadURL
			response.Shasum = p.ShaSUM
			response.OS = p.OS
			response.Arch = p.Architecture

			return response, nil
		}
	}

	return nil, nil
}

func (b *BadgerDBBackend) GetProviderVersions(_ context.Context, parameters registrytypes.ProviderVersionParameters) (*models.TerraformAvailableProvider, error) {
	opts := badger.DefaultOptions(b.DBPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer func(client *badger.DB) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	providerName := fmt.Sprintf("%s:%s/%s", b.ProviderTableName, parameters.Namespace, parameters.Name)
	prefix := []byte(providerName + ":") // Prefix for filtering

	var providerVersions []ProviderVersion
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var provider ProviderVersion
				if err := json.Unmarshal(v, &provider); err != nil {
					return err
				}
				providerVersions = append(providerVersions, provider)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	var versions []models.TerraformAvailableVersion
	for _, pv := range providerVersions {
		v := models.TerraformAvailableVersion{
			Version:   pv.Version,
			Protocols: pv.Protocols,
		}

		for _, p := range pv.Provider {
			platform := models.TerraformAvailablePlatform{
				OS:   p.OS,
				Arch: p.Architecture,
			}
			v.Platforms = append(v.Platforms, platform)
		}
		versions = append(versions, v)
	}

	provider := &models.TerraformAvailableProvider{
		Versions: versions,
	}

	return provider, nil
}

func (b *BadgerDBBackend) GetModuleVersions(_ context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	opts := badger.DefaultOptions(b.DBPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer func(client *badger.DB) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	moduleName := fmt.Sprintf("%s/%s/%s", parameters.Namespace, parameters.Name, parameters.System)
	prefix := []byte(fmt.Sprintf("%s:%s:", b.ModuleTableName, moduleName)) // Prefix for filtering

	var versions []models.TerraformAvailableModuleVersion
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var module ModuleVersion
				if err := json.Unmarshal(v, &module); err != nil {
					return err
				}
				versions = append(versions, models.TerraformAvailableModuleVersion{
					Version: module.Version,
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	modules := &models.TerraformAvailableModule{
		Modules: []models.TerraformAvailableModuleVersions{
			{
				Versions: versions,
			},
		},
	}

	return modules, nil
}

func (b *BadgerDBBackend) GetModuleDownload(_ context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
	opts := badger.DefaultOptions(b.DBPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer func(client *badger.DB) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	moduleName := fmt.Sprintf("%s/%s/%s", parameters.Namespace, parameters.Name, parameters.System)
	key := []byte(fmt.Sprintf("%s:%s:%s", b.ModuleTableName, moduleName, parameters.Version)) // Prefix for filtering

	var moduleVersion ModuleVersion
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &moduleVersion)
		})
	})
	if err != nil {
		return nil, err
	}

	if moduleVersion.DownloadURL != "" {
		return &moduleVersion.DownloadURL, nil
	}

	return nil, nil
}

func (b *BadgerDBBackend) ImportProvider(_ context.Context, provider registrytypes.ProviderImport) error {
	pv := ProviderVersion{
		Version:        provider.Version,
		Name:           provider.Name,
		Protocols:      provider.Protocols,
		SHASUMUrl:      provider.SHASUMUrl,
		SHASUMSigUrl:   provider.SHASUMSigUrl,
		GPGASCIIArmor:  provider.GPGASCIIArmor,
		GPGFingerprint: provider.GPGFingerprint,
	}

	var providers []Provider
	for _, r := range provider.Release {
		p := Provider{
			OS:           r.OS,
			Architecture: r.Architecture,
			Filename:     r.Filename,
			DownloadURL:  r.DownloadUrl,
			ShaSUM:       r.SHASUM,
		}

		providers = append(providers, p)
	}
	pv.Provider = providers

	opts := badger.DefaultOptions(b.DBPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	defer func(client *badger.DB) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	key := fmt.Sprintf("%s:%s:%s", b.ProviderTableName, provider.Name, provider.Version)
	err = providerSet(db, key, pv)
	if err != nil {
		return err
	}

	return nil
}

func (b *BadgerDBBackend) ImportModule(ctx context.Context, module registrytypes.ModuleImport) error {
	opts := badger.DefaultOptions(b.DBPath).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	defer func(client *badger.DB) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	key := fmt.Sprintf("%s:%s:%s", b.ModuleTableName, module.Name, module.Version)
	mv := ModuleVersion{
		DownloadURL: module.DownloadUrl,
		Version:     module.Version,
	}
	err = moduleSet(db, key, mv)
	if err != nil {
		return err
	}

	return nil
}

func providerSet(db *badger.DB, key string, value ProviderVersion) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func moduleSet(db *badger.DB, key string, value ModuleVersion) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}
