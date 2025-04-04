package models

type ProviderVersionPlatformsRequest struct {
	Data ProviderVersionPlatformsDataRequest `json:"data"`
}

type ProviderVersionPlatformsDataRequest struct {
	Type       string                                    `json:"type"`
	Attributes ProviderVersionPlatformsAttributesRequest `json:"attributes"`
}

type ProviderVersionPlatformsAttributesRequest struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Shasum   string `json:"shasum"`
	Filename string `json:"filename"`
}
