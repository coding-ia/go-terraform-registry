package models

import "time"

// RegistryProviderVersionsRequest - Request Model
type RegistryProviderVersionsRequest struct {
	Data RegistryProviderVersionsRequestData `json:"data"`
}

type RegistryProviderVersionsRequestAttributes struct {
	Version   string   `json:"version"`
	KeyID     string   `json:"key-id"`
	Protocols []string `json:"protocols"`
}

type RegistryProviderVersionsRequestData struct {
	Type       string                                    `json:"type"`
	Attributes RegistryProviderVersionsRequestAttributes `json:"attributes"`
}

// RegistryProviderVersionsResponse - Response Model
type RegistryProviderVersionsResponse struct {
	Data RegistryProviderVersionsResponseData `json:"data"`
}

type RegistryProviderVersionsResponseAttributes struct {
	Version            string                      `json:"version"`
	CreatedAt          time.Time                   `json:"created-at"`
	UpdatedAt          time.Time                   `json:"updated-at"`
	KeyID              string                      `json:"key-id"`
	Protocols          []string                    `json:"protocols"`
	Permissions        RegistryProviderPermissions `json:"permissions"`
	ShasumsUploaded    bool                        `json:"shasums-uploaded"`
	ShasumsSigUploaded bool                        `json:"shasums-sig-uploaded"`
}

type RegistryProviderRelation struct {
	Data RegistryProviderDataRef `json:"data"`
}

type RegistryProviderDataRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type RegistryProviderPlatformLinks struct {
	Related string `json:"related"`
}

type RegistryProviderPlatforms struct {
	Data  []RegistryProviderDataRef     `json:"data"`
	Links RegistryProviderPlatformLinks `json:"links"`
}

type RegistryProviderVersionsResponseRelationships struct {
	RegistryProvider          RegistryProviderRelation  `json:"registry-provider"`
	RegistryProviderPlatforms RegistryProviderPlatforms `json:"registry-provider-platforms"`
}

type RegistryProviderVersionsResponseLinks struct {
	ShasumsUpload    string `json:"shasums-upload,omitempty"`
	ShasumsSigUpload string `json:"shasums-sig-upload,omitempty"`
}

type RegistryProviderVersionsResponseData struct {
	ID            string                                        `json:"id"`
	Type          string                                        `json:"type"`
	Attributes    RegistryProviderVersionsResponseAttributes    `json:"attributes"`
	Relationships RegistryProviderVersionsResponseRelationships `json:"relationships"`
	Links         RegistryProviderVersionsResponseLinks         `json:"links"`
}

// RegistryProviderVersionPlatformsRequest - Request Model
type RegistryProviderVersionPlatformsRequest struct {
	Data RegistryProviderVersionPlatformsRequestData `json:"data"`
}

type RegistryProviderVersionPlatformsRequestData struct {
	Type       string                                            `json:"type"`
	Attributes RegistryProviderVersionPlatformsRequestAttributes `json:"attributes"`
}

type RegistryProviderVersionPlatformsRequestAttributes struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Shasum   string `json:"shasum"`
	Filename string `json:"filename"`
}

// RegistryProviderVersionPlatformsResponse - Response Model
type RegistryProviderVersionPlatformsResponse struct {
	Data RegistryProviderVersionPlatformsResponseData `json:"data"`
}

type RegistryProviderVersionPlatformsLinks struct {
	ProviderBinaryUpload string `json:"provider-binary-upload"`
}

type ProviderVersion struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type RegistryProviderVersionPlatformsRelationships struct {
	RegistryProviderVersion struct {
		Data ProviderVersion `json:"data"`
	} `json:"registry-provider-version"`
}

type RegistryProviderVersionPlatformsResponseData struct {
	ID            string                                             `json:"id"`
	Type          string                                             `json:"type"`
	Attributes    RegistryProviderVersionPlatformsResponseAttributes `json:"attributes"`
	Relationships RegistryProviderVersionPlatformsRelationships      `json:"relationships"`
	Links         RegistryProviderVersionPlatformsLinks              `json:"links"`
}

type RegistryProviderVersionPlatformsResponseAttributes struct {
	OS                     string                      `json:"os"`
	Arch                   string                      `json:"arch"`
	Filename               string                      `json:"filename"`
	Shasum                 string                      `json:"shasum"`
	Permissions            RegistryProviderPermissions `json:"permissions"`
	ProviderBinaryUploaded bool                        `json:"provider-binary-uploaded"`
}

// Shared
type RegistryProviderPermissions struct {
	CanDelete      bool `json:"can-delete"`
	CanUploadAsset bool `json:"can-upload-asset"`
}

type RegistryProviderLinks struct {
	Self string `json:"self"`
}
