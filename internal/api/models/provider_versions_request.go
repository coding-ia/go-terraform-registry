package models

type ProviderVersionsRequest struct {
	Data ProviderVersionsDataRequest `json:"data"`
}

type ProviderVersionsDataRequest struct {
	Type       string                            `json:"type"`
	Attributes ProviderVersionsAttributesRequest `json:"attributes"`
}

type ProviderVersionsAttributesRequest struct {
	Version   string   `json:"version"`
	KeyID     string   `json:"key-id"`
	Protocols []string `json:"protocols"`
}
