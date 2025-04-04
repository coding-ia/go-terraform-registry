package models

type ProviderVersionsResponse struct {
	Data ProviderVersionsDataResponse `json:"data"`
}

type ProviderVersionsDataResponse struct {
	ID            string                                `json:"id"`
	Type          string                                `json:"type"`
	Attributes    ProviderVersionsAttributesResponse    `json:"attributes"`
	Relationships ProviderVersionsRelationshipsResponse `json:"relationships"`
	Links         ProviderVersionsLinksResponse         `json:"links"`
}

type ProviderVersionsAttributesResponse struct {
	Version            string                              `json:"version"`
	CreatedAt          string                              `json:"created-at"`
	UpdatedAt          string                              `json:"updated-at"`
	KeyID              string                              `json:"key-id"`
	Protocols          []string                            `json:"protocols"`
	Permissions        ProviderVersionsPermissionsResponse `json:"permissions"`
	ShasumsUploaded    bool                                `json:"shasums-uploaded"`
	ShasumsSigUploaded bool                                `json:"shasums-sig-uploaded"`
}

type ProviderVersionsPermissionsResponse struct {
	CanDelete      bool `json:"can-delete"`
	CanUploadAsset bool `json:"can-upload-asset"`
}

type ProviderVersionsRelationshipsResponse struct {
	RegistryProvider ProviderVersionsRelationshipSingleResponse `json:"registry-provider"`
	Platforms        ProviderVersionsRelationshipListResponse   `json:"platforms"`
}

type ProviderVersionsRelationshipSingleResponse struct {
	Data ProviderVersionsRelationshipDataResponse `json:"data"`
}

type ProviderVersionsRelationshipListResponse struct {
	Data  []ProviderVersionsRelationshipDataResponse `json:"data"`
	Links ProviderVersionsRelatedLinkResponse        `json:"links"`
}

type ProviderVersionsRelationshipDataResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ProviderVersionsRelatedLinkResponse struct {
	Related string `json:"related"`
}

type ProviderVersionsLinksResponse struct {
	ShasumsDownload    *string `json:"shasums-download,omitempty"`
	ShasumsSigDownload *string `json:"shasums-sig-download,omitempty"`
	ShasumsUpload      *string `json:"shasums-upload,omitempty"`
	ShasumsSigUpload   *string `json:"shasums-sig-upload,omitempty"`
}
