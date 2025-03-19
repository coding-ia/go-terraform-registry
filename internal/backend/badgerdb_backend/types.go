package badgerdb_backend

type Provider struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Filename     string `json:"filename"`
	DownloadURL  string `json:"download_url"`
	ShaSUM       string `json:"sha_sum"`
}

type ProviderVersion struct {
	Version        string     `json:"version"`
	Name           string     `json:"name"`
	Protocols      []string   `json:"protocols"`
	SHASUMUrl      string     `json:"shasums_url"`
	SHASUMSigUrl   string     `json:"shasums_signature_url"`
	Provider       []Provider `json:"provider"`
	GPGASCIIArmor  string     `json:"gpg_ascii_armor"`
	GPGFingerprint string     `json:"gpg_fingerprint"`
}

type ModuleVersion struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
}