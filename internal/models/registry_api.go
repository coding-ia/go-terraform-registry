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

type RegistryProvidersResponseAttributes struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	CreatedAt    time.Time `json:"created-at"`
	UpdatedAt    time.Time `json:"updated-at"`
	RegistryName string    `json:"registry-name"`
	Permissions  struct {
		CanDelete bool `json:"can-delete"`
	} `json:"permissions"`
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

// RegistryProviderVersionRequest - Request Model
type RegistryProviderVersionRequest struct {
	Data RegistryProviderVersionRequestData `json:"data"`
}

type RegistryProviderVersionRequestAttributes struct {
	Version   string   `json:"version"`
	KeyID     string   `json:"key-id"`
	Protocols []string `json:"protocols"`
}

type RegistryProviderVersionRequestData struct {
	Type       string                                   `json:"type"`
	Attributes RegistryProviderVersionRequestAttributes `json:"attributes"`
}

// RegistryProviderVersionResponse - Response Model
type RegistryProviderVersionResponse struct {
	Data RegistryProviderVersionResponseData `json:"data"`
}

type RegistryProviderVersionResponseAttributes struct {
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

type RegistryProviderVersionResponseRelationships struct {
	RegistryProvider          RegistryProviderRelation  `json:"registry-provider"`
	RegistryProviderPlatforms RegistryProviderPlatforms `json:"registry-provider-platforms"`
}

type RegistryProviderVersionResponseData struct {
	ID            string                                       `json:"id"`
	Type          string                                       `json:"type"`
	Attributes    RegistryProviderVersionResponseAttributes    `json:"attributes"`
	Relationships RegistryProviderVersionResponseRelationships `json:"relationships"`
	Links         RegistryProviderLinks                        `json:"links"`
}

// RegistryProviderVersionPlatformRequest - Request Model
type RegistryProviderVersionPlatformRequest struct {
	Data RegistryProviderVersionPlatformRequestData `json:"data"`
}

type RegistryProviderVersionPlatformRequestData struct {
	Type       string                                           `json:"type"`
	Attributes RegistryProviderVersionPlatformRequestAttributes `json:"attributes"`
}

type RegistryProviderVersionPlatformRequestAttributes struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Shasum   string `json:"shasum"`
	Filename string `json:"filename"`
}

type RegistryProviderVersionPlatformResponse struct {
	Data RegistryProviderVersionPlatformResponseData `json:"data"`
}

type RegistryProviderVersionPlatformLinks struct {
	ProviderBinaryUpload string `json:"provider-binary-upload"`
}

type ProviderVersion struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type RegistryProviderVersionPlatformRelationships struct {
	RegistryProviderVersion struct {
		Data ProviderVersion `json:"data"`
	} `json:"registry-provider-version"`
}

type RegistryProviderVersionPlatformResponseData struct {
	ID            string                                            `json:"id"`
	Type          string                                            `json:"type"`
	Attributes    RegistryProviderVersionPlatformResponseAttributes `json:"attributes"`
	Relationships RegistryProviderVersionPlatformRelationships      `json:"relationships"`
	Links         RegistryProviderVersionPlatformLinks              `json:"links"`
}

type RegistryProviderVersionPlatformResponseAttributes struct {
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
