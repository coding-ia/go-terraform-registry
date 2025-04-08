package postgres_backend

type GPGKey struct {
	Namespace  string `json:"namespace"`
	KeyID      string `json:"key_id"`
	ID         string `json:"id"`
	AsciiArmor string `json:"ascii_armor"`
}

type Provider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Organization string `json:"organization"`
	RegistryName string `json:"registry_name"`
}

type Module struct {
	ID           string `json:"id"`
	Provider     string `json:"provider"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Organization string `json:"organization"`
	RegistryName string `json:"registry_name"`
	NoCode       bool   `json:"no_code"`
}

type ModuleVersion struct {
	ID        string `json:"id"`
	ModuleID  string `json:"module_id"`
	Version   string `json:"version"`
	CommitSHA string `json:"commit_sha"`
}

type ProviderVersion struct {
	ID         string                  `json:"id"`
	ProviderID string                  `json:"provider_id"`
	GPGKeyID   string                  `json:"gpg_key_id"`
	Version    string                  `json:"version"`
	MetaData   ProviderVersionMetaData `json:"metadata"`
}

type ProviderVersionMetaData struct {
	Protocols []string `json:"protocols"`
}

type ProviderPlatform struct {
	ID                string `json:"id"`
	OS                string `json:"os"`
	Arch              string `json:"arch"`
	SHASum            string `json:"shasum"`
	Filename          string `json:"filename"`
	ProviderVersionID string `json:"provider_version_id"`
}

type ProviderRelease struct {
	Organization  string             `json:"organization"`
	Repository    string             `json:"repository"`
	Namespace     string             `json:"namespace"`
	Name          string             `json:"name"`
	Version       string             `json:"version"`
	Protocols     []string           `json:"metadata"`
	Platforms     []ProviderPlatform `json:"platforms"`
	GPGASCIIArmor string             `json:"gpg_ascii_armor"`
	GPGKeyID      string             `json:"gpg_key_id"`
}

type ModuleRelease struct {
	Organization string                 `json:"organization"`
	Registry     string                 `json:"registry"`
	Namespace    string                 `json:"namespace"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Versions     []ModuleReleaseVersion `json:"versions"`
}

type ModuleReleaseVersion struct {
	Version string `json:"version"`
}

type ModuleReleaseVersionInfo struct {
	Organization string `json:"organization"`
	Registry     string `json:"registry"`
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	Version      string `json:"version"`
	CommitSHA    string `json:"commit_sha"`
}
