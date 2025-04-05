PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS gpg_keys (
    id TEXT PRIMARY KEY NOT NULL,
    type TEXT NOT NULL,
    ascii_armor TEXT NOT NULL,
    created_at TEXT NOT NULL,
    key_id TEXT NOT NULL,
    namespace TEXT NOT NULL,
    source TEXT,
    source_url TEXT,
    trust_signature TEXT,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS providers (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    namespace TEXT NOT NULL,
	organization TEXT NOT NULL,
    registry_name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    permissions TEXT
);

CREATE TABLE IF NOT EXISTS provider_versions (
    id TEXT PRIMARY KEY NOT NULL,
    provider_id TEXT NOT NULL,
    version TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    key_id TEXT NOT NULL,
    protocols TEXT NOT NULL,
    permissions TEXT,
    FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
    FOREIGN KEY (key_id) REFERENCES gpg_keys(id)
);

CREATE TABLE IF NOT EXISTS provider_version_platforms (
    id TEXT PRIMARY KEY NOT NULL,
    provider_version_id TEXT NOT NULL,
    os TEXT NOT NULL,
    arch TEXT NOT NULL,
    filename TEXT NOT NULL,
    shasum TEXT NOT NULL,
    permissions TEXT NOT NULL,
    provider_binary_uploaded BOOLEAN NOT NULL,
    FOREIGN KEY (provider_version_id) REFERENCES provider_versions(id) ON DELETE CASCADE
);
