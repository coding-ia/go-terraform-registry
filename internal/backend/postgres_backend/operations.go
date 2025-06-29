package postgres_backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func WithTransaction(ctx context.Context, db *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func gpgInsert(ctx context.Context, db *pgxpool.Pool, value GPGKey) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO gpg_keys (ascii_armor, key_id, namespace)
			VALUES ($1, $2, $3)
	`
		_, err := tx.Exec(ctx, query, value.AsciiArmor, value.KeyID, value.Namespace)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "unique_keyid_namespace" {
					return fmt.Errorf("gpg key already exists")
				}
			}
		}
		return err
	})
}

func gpgSelect(ctx context.Context, db *pgxpool.Pool, keyID, namespace string) (*GPGKey, error) {
	query := `
		SELECT gpgkey_id, ascii_armor, key_id, namespace
		FROM gpg_keys
		WHERE key_id = $1 AND namespace = $2;
	`

	row := db.QueryRow(ctx, query, keyID, namespace)

	var key GPGKey
	err := row.Scan(
		&key.ID,
		&key.AsciiArmor,
		&key.KeyID,
		&key.Namespace,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &key, nil
}

func gpgList(ctx context.Context, db *pgxpool.Pool, namespace string) (*[]GPGKey, *Pagination, error) {
	queryCount := `
		SELECT COUNT(*)
		FROM gpg_keys
		WHERE namespace = $1;
	`

	query := `
		SELECT gpgkey_id, ascii_armor, key_id, namespace
		FROM gpg_keys
		WHERE namespace = $1;
	`

	var pagination Pagination
	err := db.QueryRow(ctx, queryCount, namespace).Scan(&pagination.TotalCount)
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.Query(ctx, query, namespace)
	if err != nil {
		return nil, nil, err
	}

	var keys []GPGKey
	for rows.Next() {
		var key GPGKey
		err := rows.Scan(
			&key.ID,
			&key.AsciiArmor,
			&key.KeyID,
			&key.Namespace,
		)
		if err != nil {
			return nil, nil, err
		}

		keys = append(keys, key)
	}

	return &keys, &pagination, nil
}

func modulesInsert(ctx context.Context, db *pgxpool.Pool, value *Module) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO modules (provider, name, namespace, organization, registry)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING module_id;
	`
		err := tx.QueryRow(ctx, query, value.Provider, value.Name, value.Namespace, value.Organization, value.RegistryName).Scan(&value.ID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "unique_module_identity" {
					return fmt.Errorf("module already exists")
				}
			}
		}
		return err
	})
}

func getModuleReleases(ctx context.Context, db *pgxpool.Pool, organization string, registry string, namespace string, name string, provider string) (*ModuleRelease, error) {
	query := `
		SELECT organization, registry, namespace, name, provider, versions
		FROM registry_modules
		WHERE organization = $1 AND registry = $2 AND namespace = $3 AND name = $4 AND provider = $5;
	`

	row := db.QueryRow(ctx, query, organization, registry, namespace, name, provider)
	if row == nil {
		return nil, fmt.Errorf("no provider release found for %s/%s/%s", namespace, name, provider)
	}

	var mr ModuleRelease
	err := row.Scan(&mr.Organization, &mr.Registry, &mr.Namespace, &mr.Name, &mr.Provider, &mr.Versions)
	if err != nil {
		return nil, err
	}

	return &mr, nil
}

func getModuleRelease(ctx context.Context, db *pgxpool.Pool, organization string, registry string, namespace string, name string, provider string, version string) (*ModuleReleaseVersionInfo, error) {
	query := `
		SELECT organization, registry, namespace, name, provider, version, commit_sha
		FROM registry_module_versions
		WHERE organization = $1 AND registry = $2 AND namespace = $3 AND name = $4 AND provider = $5 AND version = $6;
	`

	row := db.QueryRow(ctx, query, organization, registry, namespace, name, provider, version)
	if row == nil {
		return nil, fmt.Errorf("no provider release found for %s/%s/%s", namespace, name, provider)
	}

	var mr ModuleReleaseVersionInfo
	err := row.Scan(&mr.Organization, &mr.Registry, &mr.Namespace, &mr.Name, &mr.Provider, &mr.Version, &mr.CommitSHA)
	if err != nil {
		return nil, err
	}

	return &mr, nil
}

func modulesSelect(ctx context.Context, db *pgxpool.Pool, organization string, registry string, namespace string, name string, provider string) (*Module, error) {
	query := `
		SELECT module_id, provider, name, namespace, organization, registry
		FROM modules
		WHERE provider = $1 AND name = $2 AND namespace = $3 AND organization = $4 AND registry = $5;
	`

	row := db.QueryRow(ctx, query, provider, name, namespace, organization, registry)

	var module Module
	err := row.Scan(
		&module.ID,
		&module.Provider,
		&module.Name,
		&module.Namespace,
		&module.Organization,
		&module.RegistryName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &module, nil
}

func moduleVersionsInsert(ctx context.Context, db *pgxpool.Pool, value *ModuleVersion) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO module_versions (module_id, version, commit_sha)
			VALUES ($1, $2, $3)
			RETURNING module_version_id;
	`
		err := tx.QueryRow(ctx, query, value.ModuleID, value.Version, value.CommitSHA).Scan(&value.ID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "unique_module_version" {
					return fmt.Errorf("module version already exists")
				}
			}
		}
		return err
	})
}

func moduleVersionsDelete(ctx context.Context, db *pgxpool.Pool, moduleId string, version string) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			DELETE FROM module_versions
			WHERE module_id = $1 AND version = $2;
	`
		_, err := tx.Exec(ctx, query, moduleId, version)
		return err
	})
}

func providersInsert(ctx context.Context, db *pgxpool.Pool, value *Provider) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO providers (name, namespace, organization, registry)
			VALUES ($1, $2, $3, $4)
			RETURNING provider_id;
	`
		err := tx.QueryRow(ctx, query, value.Name, value.Namespace, value.Organization, value.RegistryName).Scan(&value.ID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "unique_provider_identity" {
					return fmt.Errorf("provider already exists")
				}
			}
		}
		return err
	})
}

func providersSelect(ctx context.Context, db *pgxpool.Pool, organization string, registry string, namespace string, name string) (*Provider, error) {
	query := `
		SELECT provider_id, name, namespace, organization, registry
		FROM providers
		WHERE name = $1 AND namespace = $2 AND organization = $3 AND registry = $4;
	`

	row := db.QueryRow(ctx, query, name, namespace, organization, registry)

	var provider Provider
	err := row.Scan(
		&provider.ID,
		&provider.Name,
		&provider.Namespace,
		&provider.Organization,
		&provider.RegistryName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &provider, nil
}

func getProviderRelease(ctx context.Context, db *pgxpool.Pool, organization string, registry string, namespace string, name string, version string) (*ProviderRelease, error) {
	query := `
		SELECT organization, registry, namespace, name, key_id, ascii_armor, version, protocols, platforms
		FROM registry_provider_release
		WHERE organization = $1 AND registry = $2 AND namespace = $3 AND name = $4 AND version = $5;

`

	row := db.QueryRow(ctx, query, organization, registry, namespace, name, version)
	if row == nil {
		return nil, fmt.Errorf("no provider release found for %s/%s", namespace, name)
	}

	var pr ProviderRelease
	err := row.Scan(&pr.Organization, &pr.Repository, &pr.Namespace, &pr.Name, &pr.GPGKeyID, &pr.GPGASCIIArmor, &pr.Version, &pr.Protocols, &pr.Platforms)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func getProviderReleases(ctx context.Context, db *pgxpool.Pool, organization string, registry string, namespace string, name string) ([]ProviderRelease, error) {
	query := `
		SELECT organization, registry, namespace, name, version, protocols, platforms
		FROM registry_provider_releases
		WHERE organization = $1 AND registry = $2 AND namespace = $3 AND name = $4;

`

	rows, err := db.Query(ctx, query, organization, registry, namespace, name)
	if err != nil {
		return nil, err
	}

	var releases []ProviderRelease
	for rows.Next() {
		var pr ProviderRelease

		err = rows.Scan(&pr.Organization, &pr.Repository, &pr.Namespace, &pr.Name, &pr.Version, &pr.Protocols, &pr.Platforms)
		if err != nil {
			return nil, err
		}

		releases = append(releases, pr)
	}

	return releases, nil
}

func providerVersionsInsert(ctx context.Context, db *pgxpool.Pool, value *ProviderVersion) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO provider_versions (provider_id, gpgkey_id, version, metadata)
			VALUES ($1, $2, $3, $4)
			RETURNING provider_version_id;
	`
		md, err := json.Marshal(value.MetaData)
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx, query, value.ProviderID, value.GPGKeyID, value.Version, string(md)).Scan(&value.ID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "unique_provider_version" {
					return fmt.Errorf("provider version already exists")
				}
			}
		}

		return err
	})
}

func providerVersionsList(ctx context.Context, db *pgxpool.Pool, providerId string) (*map[string]ProviderVersion, *Pagination, error) {
	queryCount := `
		SELECT COUNT(*)
		FROM provider_versions
		WHERE provider_id = $1;
	`

	query := `
		WITH limited_versions AS (
		  SELECT *
		  FROM provider_versions
		  WHERE provider_id = $1
		  ORDER BY created_at DESC
		)
		SELECT
		  pv.provider_version_id,
		  pv.provider_id,
		  pv.gpgkey_id,
		  pv.version,
          pv.metadata,
          pvp.provider_version_platform_id
		FROM limited_versions pv
		LEFT JOIN provider_version_platforms pvp
		  ON pv.provider_version_id = pvp.provider_version_id
		ORDER BY pv.created_at DESC;
	`

	var pagination Pagination
	err := db.QueryRow(ctx, queryCount, providerId).Scan(&pagination.TotalCount)
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.Query(ctx, query, providerId)
	if err != nil {
		return nil, nil, err
	}

	versions := make(map[string]ProviderVersion)
	for rows.Next() {
		var providerVersion ProviderVersion
		var metaData string
		var platformID string
		err := rows.Scan(
			&providerVersion.ID,
			&providerVersion.ProviderID,
			&providerVersion.GPGKeyID,
			&providerVersion.Version,
			&metaData,
			&platformID,
		)
		if err != nil {
			return nil, nil, err
		}

		if _, exists := versions[providerVersion.ID]; !exists {
			err = json.Unmarshal([]byte(metaData), &providerVersion.MetaData)
			if err != nil {
				return nil, nil, err
			}
			providerVersion.Platforms = append(providerVersion.Platforms, platformID)
			versions[providerVersion.ID] = providerVersion
		} else {
			tmpVer := versions[providerVersion.ID]
			tmpVer.Platforms = append(tmpVer.Platforms, platformID)
			versions[providerVersion.ID] = tmpVer
		}
	}

	return &versions, &pagination, nil
}

func providerVersionSelect(ctx context.Context, db *pgxpool.Pool, providerId string, version string) (*ProviderVersion, error) {
	query := `
		SELECT provider_version_id, provider_id, gpgkey_id, version, metadata
		FROM provider_versions
		WHERE provider_id = $1 AND version = $2;
	`

	row := db.QueryRow(ctx, query, providerId, version)

	var providerVersion ProviderVersion
	var metaData string
	err := row.Scan(
		&providerVersion.ID,
		&providerVersion.ProviderID,
		&providerVersion.GPGKeyID,
		&providerVersion.Version,
		&metaData,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(metaData), &providerVersion.MetaData)
	if err != nil {
		return nil, err
	}

	return &providerVersion, nil
}

func providerVersionDelete(ctx context.Context, db *pgxpool.Pool, providerId string, version string) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			DELETE FROM provider_versions
			WHERE provider_id = $1 AND version = $2;
	`
		_, err := tx.Exec(ctx, query, providerId, version)
		return err
	})
}

func providerVersionPlatformInsert(ctx context.Context, db *pgxpool.Pool, value *ProviderPlatform) error {
	return WithTransaction(ctx, db, func(tx pgx.Tx) error {
		query := `
			INSERT INTO provider_version_platforms (provider_version_id, os, arch, filename, shasum, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING provider_version_platform_id;
	`
		err := tx.QueryRow(ctx, query, value.ProviderVersionID, value.OS, value.Arch, value.Filename, value.SHASum, "{}").Scan(&value.ID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "unique_provider_version_platform" {
					return fmt.Errorf("platform already exists")
				}
			}
		}
		return err
	})
}
