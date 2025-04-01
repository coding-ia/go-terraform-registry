package badgerdb_backend

type Provider struct {
	ID string `json:"id"`
}

type ProviderVersion struct {
	ID            string             `json:"id"`
	Version       string             `json:"version"`
	Platform      []ProviderPlatform `json:"platform"`
	Protocols     []string           `json:"protocols"`
	GPGASCIIArmor string             `json:"gpg_ascii_armor"`
	GPGKeyID      string             `json:"gpg_key_id"`
}

type ProviderPlatform struct {
	ID       string `json:"id"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	SHASum   string `json:"shasum"`
	Filename string `json:"filename"`
}
