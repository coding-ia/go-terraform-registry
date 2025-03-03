package models

type TerraformAvailablePlatform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type TerraformAvailableVersion struct {
	Version   string                       `json:"version"`
	Protocols []string                     `json:"protocols"`
	Platforms []TerraformAvailablePlatform `json:"platforms"`
}

type TerraformAvailableProvider struct {
	Versions []TerraformAvailableVersion `json:"versions"`
}

type GPGPublicKeys struct {
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceUrl      string `json:"source_url"`
}

type SigningKeys struct {
	GPGPublicKeys []GPGPublicKeys `json:"gpg_public_keys"`
}

type TerraformProviderPlatformResponse struct {
	Protocols           []string    `json:"protocols"`
	OS                  string      `json:"os"`
	Arch                string      `json:"arch"`
	Filename            string      `json:"filename"`
	DownloadUrl         string      `json:"download_url"`
	ShasumsUrl          string      `json:"shasums_url"`
	ShasumsSignatureUrl string      `json:"shasums_signature_url"`
	Shasum              string      `json:"shasum"`
	SigningKeys         SigningKeys `json:"signing_keys"`
}
