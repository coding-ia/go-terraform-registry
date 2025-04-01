package badgerdb_backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/models"
	registrytypes "go-terraform-registry/internal/types"
	"golang.org/x/crypto/openpgp"
	"log"
	"os"
	"strings"
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

func (b *BadgerDBBackend) GetProvider(ctx context.Context, parameters registrytypes.ProviderPackageParameters, userParameters registrytypes.UserParameters) (*models.TerraformProviderPlatformResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.ProviderTableName, userParameters.Organization, "private", parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("%s:%s:%s", b.ProviderVersionTableName, p.ID, parameters.Version)

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
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.ProviderTableName, userParameters.Organization, "private", parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("%s:%s", b.ProviderVersionTableName, p.ID)
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

func (b *BadgerDBBackend) RegistryProviders(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProvidersRequest) (*models.RegistryProvidersResponse, error) {
	newUUID := uuid.New()

	p := Provider{
		ID: newUUID.String(),
	}

	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.ProviderTableName, parameters.Organization, request.Data.Attributes.RegistryName, request.Data.Attributes.Namespace, request.Data.Attributes.Name)
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerSet(db, key, p)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.RegistryProvidersResponse{
		Data: models.RegistryProvidersResponseData{
			ID:   p.ID,
			Type: "registry-providers",
			Attributes: models.RegistryProvidersResponseAttributes{
				Name:         request.Data.Attributes.Name,
				Namespace:    request.Data.Attributes.Namespace,
				RegistryName: request.Data.Attributes.RegistryName,
				Permissions: models.RegistryProvidersResponsePermissions{
					CanDelete: true,
				},
			},
		},
	}

	return resp, nil
}

func (b *BadgerDBBackend) GPGKey(ctx context.Context, request models.GPGKeyRequest) (*models.GPGKeyResponse, error) {
	fingerprint := getKeyFingerprint(request.Data.Attributes.AsciiArmor)

	key := fmt.Sprintf("%s:%s:%s", b.GPGTableName, request.Data.Attributes.Namespace, fingerprint[0])
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return gpgSet(db, key, request.Data.Attributes.AsciiArmor)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.GPGKeyResponse{
		Data: models.GPGKeyResponseData{
			ID: fingerprint[0],
			Attributes: models.GPGKeyResponseAttributes{
				AsciiArmor: request.Data.Attributes.AsciiArmor,
				KeyID:      fingerprint[0],
				Namespace:  request.Data.Attributes.Namespace,
			},
		},
	}

	return resp, nil
}

func (b *BadgerDBBackend) RegistryProviderVersions(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionsRequest) (*models.RegistryProviderVersionsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.ProviderTableName, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	var asciiArmor string
	gpgKey := fmt.Sprintf("%s:%s:%s", b.GPGTableName, parameters.Namespace, request.Data.Attributes.KeyID)
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return gpgGet(db, gpgKey, &asciiArmor)
	})
	if err != nil {
		return nil, err
	}

	newUUID := uuid.New()
	pv := ProviderVersion{
		ID:            newUUID.String(),
		Version:       request.Data.Attributes.Version,
		Protocols:     request.Data.Attributes.Protocols,
		GPGKeyID:      request.Data.Attributes.KeyID,
		GPGASCIIArmor: asciiArmor,
	}

	pvKey := fmt.Sprintf("%s:%s:%s", b.ProviderVersionTableName, p.ID, pv.Version)
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionSet(db, pvKey, pv)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.RegistryProviderVersionsResponse{
		Data: models.RegistryProviderVersionsResponseData{
			ID:   pv.ID,
			Type: "registry-provider-versions",
			Attributes: models.RegistryProviderVersionsResponseAttributes{
				Version:   pv.Version,
				Protocols: pv.Protocols,
				KeyID:     pv.GPGKeyID,
			},
		},
	}

	return resp, nil
}

func (b *BadgerDBBackend) RegistryProviderVersionPlatforms(ctx context.Context, parameters registrytypes.APIParameters, request models.RegistryProviderVersionPlatformsRequest) (*models.RegistryProviderVersionPlatformsResponse, error) {
	key := fmt.Sprintf("%s:%s:%s:%s/%s", b.ProviderTableName, parameters.Organization, parameters.Registry, parameters.Namespace, parameters.Name)

	var p Provider
	err := withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerGet(db, key, &p)
	})
	if err != nil {
		return nil, err
	}

	pvKey := fmt.Sprintf("%s:%s:%s", b.ProviderVersionTableName, p.ID, parameters.Version)
	var pv ProviderVersion
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionGet(db, pvKey, &pv)
	})
	if err != nil {
		return nil, err
	}

	newUUID := uuid.New()
	platform := ProviderPlatform{
		ID:       newUUID.String(),
		OS:       request.Data.Attributes.OS,
		Arch:     request.Data.Attributes.Arch,
		SHASum:   request.Data.Attributes.Shasum,
		Filename: request.Data.Attributes.Filename,
	}

	pv.Platform = append(pv.Platform, platform)
	err = withBadgerDB(b.DBPath, func(db *badger.DB) error {
		return providerVersionSet(db, pvKey, pv)
	})
	if err != nil {
		return nil, err
	}

	resp := &models.RegistryProviderVersionPlatformsResponse{
		Data: models.RegistryProviderVersionPlatformsResponseData{
			ID:   platform.ID,
			Type: "registry-provider-platforms",
			Attributes: models.RegistryProviderVersionPlatformsResponseAttributes{
				OS:       platform.OS,
				Arch:     platform.Arch,
				Shasum:   platform.SHASum,
				Filename: platform.Filename,
			},
		},
	}

	return resp, nil
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
			return err
		}

		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &value)
		})
	})
}

func gpgSet(db *badger.DB, key string, value string) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func gpgGet(db *badger.DB, key string, value *string) error {
	return db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(v []byte) error {
			*value = string(v)
			return nil
		})
	})
}

func getKeyFingerprint(publicKey string) []string {
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(publicKey))
	if err != nil {
		log.Fatal(err)
	}

	var keys []string
	for _, entity := range entityList {
		fingerPrint := entity.PrimaryKey.Fingerprint
		keyID := fingerPrint[len(fingerPrint)-8:]
		value := fmt.Sprintf("%x", keyID)
		keys = append(keys, strings.ToUpper(value))
	}

	return keys
}
