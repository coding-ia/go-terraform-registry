package models

type ProviderManifest struct {
	Version  int      `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	ProtocolVersions []string `json:"protocol_versions"`
}
