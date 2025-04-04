package models

type ProviderVersionPlatformsResponse struct {
	Data ProviderVersionPlatformsDataResponse `json:"data"`
}

type ProviderVersionPlatformsDataResponse struct {
	ID            string                                        `json:"id"`
	Type          string                                        `json:"type"`
	Attributes    ProviderVersionPlatformsAttributesResponse    `json:"attributes"`
	Relationships ProviderVersionPlatformsRelationshipsResponse `json:"relationships"`
	Links         ProviderVersionPlatformsLinksResponse         `json:"links"`
}

type ProviderVersionPlatformsAttributesResponse struct {
	OS                     string                                      `json:"os"`
	Arch                   string                                      `json:"arch"`
	Filename               string                                      `json:"filename"`
	Shasum                 string                                      `json:"shasum"`
	Permissions            ProviderVersionPlatformsPermissionsResponse `json:"permissions"`
	ProviderBinaryUploaded bool                                        `json:"provider-binary-uploaded"`
}

type ProviderVersionPlatformsPermissionsResponse struct {
	CanDelete      bool `json:"can-delete"`
	CanUploadAsset bool `json:"can-upload-asset"`
}

type ProviderVersionPlatformsRelationshipsResponse struct {
	RegistryProviderVersion ProviderVersionPlatformsRegistryVersionResponse `json:"registry-provider-version"`
}

type ProviderVersionPlatformsRegistryVersionResponse struct {
	Data ProviderVersionPlatformsRegistryVersionDataResponse `json:"data"`
}

type ProviderVersionPlatformsRegistryVersionDataResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ProviderVersionPlatformsLinksResponse struct {
	ProviderBinaryUpload string `json:"provider-binary-upload"`
}
