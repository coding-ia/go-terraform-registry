package dynamodb_backend

type Provider struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
}

type GPGKey struct {
	Namespace  string `json:"namespace"`
	KeyID      string `json:"key_id"`
	ID         string `json:"id"`
	AsciiArmor string `json:"ascii_armor"`
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
