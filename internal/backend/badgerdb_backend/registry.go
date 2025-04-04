package badgerdb_backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"log"
	"strings"
)

var _ backend.RegistryBackend = &BadgerDBBackend{}

func (b *BadgerDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.Tables.ProviderTableName, userParameters.Organization, "private", parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("%s:%s:%s", b.Tables.ProviderVersionTableName, p.ID, parameters.Version)

	var pv ProviderVersion
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionGet(db, filter, &pv)
	})
	if err != nil {
		return nil, err
	}

	response := &models.TerraformProviderPlatformResponse{
		Protocols: pv.Protocols,
		SigningKeys: models.SigningKeys{
			GPGPublicKeys: []models.GPGPublicKeys{
				{
					KeyId:      pv.GPGKeyID,
					AsciiArmor: pv.GPGASCIIArmor,
				},
			},
		},
	}

	for _, platform := range pv.Platform {
		if strings.EqualFold(platform.OS, parameters.OS) &&
			strings.EqualFold(platform.Arch, parameters.Architecture) {
			response.Filename = platform.Filename
			response.Shasum = platform.SHASum
			response.OS = platform.OS
			response.Arch = platform.Arch

			return response, nil
		}
	}

	return nil, nil
}

func (b *BadgerDBBackend) GetProviderVersions(ctx context.Context, parameters registrytypes.ProviderVersionParameters, userParameters registrytypes.UserParameters) (*models.TerraformAvailableProvider, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.Tables.ProviderTableName, userParameters.Organization, "private", parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("%s:%s", b.Tables.ProviderVersionTableName, p.ID)
	prefix := []byte(filter + ":") // Prefix for filtering

	var providerVersions []ProviderVersion
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return db.View(func(txn *badger.Txn) error {
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
	})
	if err != nil {
		return nil, err
	}

	var versions []models.TerraformAvailableVersion
	for _, pv := range providerVersions {
		v := models.TerraformAvailableVersion{
			Version:   pv.Version,
			Protocols: pv.Protocols,
		}

		for _, p := range pv.Platform {
			platform := models.TerraformAvailablePlatform{
				OS:   p.OS,
				Arch: p.Arch,
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

func (b *BadgerDBBackend) GetModuleVersions(ctx context.Context, parameters registrytypes.ModuleVersionParameters) (*models.TerraformAvailableModule, error) {
	return nil, nil
}

func (b *BadgerDBBackend) GetModuleDownload(ctx context.Context, parameters registrytypes.ModuleDownloadParameters) (*string, error) {
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

func providerSet(db *badger.DB, key string, value Provider) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func providerGet(db *badger.DB, key string, value *Provider) error {
	return db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return err
		}

		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &value)
		})
	})
}

func providerVersionSet(db *badger.DB, key string, value ProviderVersion) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func providerVersionGet(db *badger.DB, key string, value *ProviderVersion) error {
	return db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return err
		}

		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &value)
		})
	})
}

func gpgSet(db *badger.DB, key string, value GPGKey) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func gpgGet(db *badger.DB, key string, value *GPGKey) error {
	return db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &value)
		})
	})
}

func duplicatePlatform(platforms []ProviderPlatform, os string, arch string) bool {
	for _, platform := range platforms {
		if strings.EqualFold(platform.OS, os) && strings.EqualFold(platform.Arch, arch) {
			return true
		}
	}

	return false
}
