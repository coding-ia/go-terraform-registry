package badgerdb_backend

type Provider struct {
	ID string `json:"id"`
}

type ProviderVersion struct {
	ID            string   `json:"id"`
	Version       string   `json:"version"`
	Protocols     []string `json:"protocols"`
	GPGASCIIArmor string   `json:"gpg_ascii_armor"`
	GPGKeyID      string   `json:"gpg_key_id"`
}
