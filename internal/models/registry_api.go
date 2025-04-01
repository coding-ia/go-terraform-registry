package models

import "time"

// RegistryProvidersRequest - Provider Request Models
type RegistryProvidersRequest struct {
	Data RegistryProvidersRequestData `json:"data"`
}

type RegistryProvidersRequestData struct {
	Type       string                             `json:"type"`
	Attributes RegistryProvidersRequestAttributes `json:"attributes"`
}

type RegistryProvidersRequestAttributes struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RegistryName string `json:"registry-name"`
}

// RegistryProvidersResponse - Provider Response Models
type RegistryProvidersResponse struct {
	Data RegistryProvidersResponseData `json:"data"`
}

type RegistryProvidersResponsePermissions struct {
	CanDelete bool `json:"can-delete"`
}

type RegistryProvidersResponseAttributes struct {
	Name         string                               `json:"name"`
	Namespace    string                               `json:"namespace"`
	CreatedAt    time.Time                            `json:"created-at"`
	UpdatedAt    time.Time                            `json:"updated-at"`
	RegistryName string                               `json:"registry-name"`
	Permissions  RegistryProvidersResponsePermissions `json:"permissions"`
}

type RegistryProvidersResponseRegistryProviderVersions struct {
	Data  []interface{} `json:"data"`
	Links struct {
		Related string `json:"related"`
	} `json:"links"`
}

type RegistryProvidersResponseOrganizationData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type RegistryProvidersResponseRelationships struct {
	Organization struct {
		Data RegistryProvidersResponseOrganizationData `json:"data"`
	} `json:"organization"`
	RegistryProviderVersions RegistryProvidersResponseRegistryProviderVersions `json:"registry-provider-versions"`
}

type RegistryProvidersResponseData struct {
	ID            string                                 `json:"id"`
	Type          string                                 `json:"type"`
	Attributes    RegistryProvidersResponseAttributes    `json:"attributes"`
	Relationships RegistryProvidersResponseRelationships `json:"relationships"`
	Links         RegistryProviderLinks                  `json:"links"`
}

// GPGKeyRequest - Request Model Data
type GPGKeyRequest struct {
	Data GPGKeyRequestData `json:"data"`
}

type GPGKeyRequestData struct {
	Type       string                  `json:"type"`
	Attributes GPGKeyRequestAttributes `json:"attributes"`
}

type GPGKeyRequestAttributes struct {
	Namespace  string `json:"namespace"`
	AsciiArmor string `json:"ascii-armor"`
}

// GPGKeyResponse - Response Model
type GPGKeyResponse struct {
	Data GPGKeyResponseData `json:"data"`
}

type GPGKeyResponseAttributes struct {
	AsciiArmor     string  `json:"ascii-armor"`
	CreatedAt      string  `json:"created-at"`
	KeyID          string  `json:"key-id"`
	Namespace      string  `json:"namespace"`
	Source         string  `json:"source"`
	SourceURL      *string `json:"source-url"`
	TrustSignature string  `json:"trust-signature"`
	UpdatedAt      string  `json:"updated-at"`
}

type GPGKeyResponseData struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Attributes GPGKeyResponseAttributes `json:"attributes"`
	Links      RegistryProviderLinks    `json:"links"`
}

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

type RegistryProviderVersionsResponseData struct {
	ID            string                                        `json:"id"`
	Type          string                                        `json:"type"`
	Attributes    RegistryProviderVersionsResponseAttributes    `json:"attributes"`
	Relationships RegistryProviderVersionsResponseRelationships `json:"relationships"`
	Links         RegistryProviderLinks                         `json:"links"`
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
