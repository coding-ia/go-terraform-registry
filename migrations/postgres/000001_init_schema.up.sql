CREATE TABLE IF NOT EXISTS gpg_keys
(
  gpgkey_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ascii_armor text NOT NULL,
  key_id varchar(40) NOT NULL,
  namespace varchar(255) NOT NULL,
  source text,
  source_url text,
  trust_signature text,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  
  CONSTRAINT unique_keyid_namespace UNIQUE (key_id, namespace)
);

CREATE TABLE IF NOT EXISTS providers
(
  provider_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(255) NOT NULL,
  namespace varchar(255) NOT NULL,
  organization varchar(255) NOT NULL,
  registry varchar(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  metadata JSONB,
  
  CONSTRAINT unique_provider_identity UNIQUE (name, namespace, organization, registry)
);

CREATE TABLE IF NOT EXISTS provider_versions
(
  provider_version_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  provider_id uuid NOT NULL,
  version varchar(32) NOT NULL,
  gpgkey_id uuid NOT NULL,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (provider_id) REFERENCES providers(provider_id) ON DELETE CASCADE,
  FOREIGN KEY (gpgkey_id) REFERENCES gpg_keys(gpgkey_id),
  
  CONSTRAINT unique_provider_version UNIQUE (provider_id, version)
);

CREATE TABLE IF NOT EXISTS provider_version_platforms
(
  provider_version_platform_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  provider_version_id uuid NOT NULL,
  os varchar(64) NOT NULL,
  arch varchar(64) NOT NULL,
  filename varchar(255) NOT NULL,
  shasum varchar(255) NOT NULL,
  metadata JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (provider_version_id) REFERENCES provider_versions(provider_version_id) ON DELETE CASCADE,
  
  CONSTRAINT unique_provider_version_platform UNIQUE (provider_version_id, os, arch)
);

CREATE TABLE IF NOT EXISTS modules
(
  module_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(255) NOT NULL,
  namespace varchar(255) NOT NULL,
  organization varchar(255) NOT NULL,
  registry varchar(255) NOT NULL,
  version varchar(32) NOT NULL,
  link varchar(2048) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  metadata JSONB,
  
  CONSTRAINT unique_module_identity UNIQUE (name, namespace, organization, registry, version)
);

CREATE INDEX IF NOT EXISTS idx_module_versions ON modules (organization, registry, namespace, name);

CREATE VIEW registry_provider_releases AS
  SELECT
    p.organization,
	p.registry,
	p.namespace,
    p.name,
    pv.version,
    (pv.metadata -> 'protocols')::json AS protocols,
    JSON_AGG(JSON_BUILD_OBJECT('os', pvp.os, 'arch', pvp.arch)) AS platforms
  FROM providers p
  JOIN provider_versions pv ON p.provider_id = pv.provider_id
  JOIN provider_version_platforms pvp ON pv.provider_version_id = pvp.provider_version_id
  GROUP BY p.organization, p.registry, p.namespace, p.name, pv.version, pv.metadata -> 'protocols';

CREATE VIEW registry_provider_release AS
  SELECT
    p.organization,
	p.registry,
	p.namespace,
    p.name,
    pv.version,
    g.key_id,
	g.ascii_armor,
    (pv.metadata -> 'protocols')::json AS protocols,
    JSON_AGG(JSON_BUILD_OBJECT('os', pvp.os, 'arch', pvp.arch, 'shasum', pvp.shasum, 'filename', pvp.filename)) AS platforms
  FROM providers p
  JOIN provider_versions pv ON p.provider_id = pv.provider_id
  JOIN gpg_keys g ON pv.gpgkey_id = g.gpgkey_id
  JOIN provider_version_platforms pvp ON pv.provider_version_id = pvp.provider_version_id
  GROUP BY p.organization, p.registry, p.namespace, p.name, g.key_id, g.ascii_armor, pv.version, pv.metadata -> 'protocols';
