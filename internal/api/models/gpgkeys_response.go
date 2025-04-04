package models

type GPGKeysResponse struct {
	Data GPGKeysDataResponse `json:"data"`
}

type GPGKeysDataResponse struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes GPGKeysAttributesResponse `json:"attributes"`
	Links      GPGKeysLinksResponse      `json:"links"`
}

type GPGKeysAttributesResponse struct {
	AsciiArmor     string  `json:"ascii-armor"`
	CreatedAt      string  `json:"created-at"`
	KeyID          string  `json:"key-id"`
	Namespace      string  `json:"namespace"`
	Source         string  `json:"source"`
	SourceURL      *string `json:"source-url"` // nullable
	TrustSignature string  `json:"trust-signature"`
	UpdatedAt      string  `json:"updated-at"`
}

type GPGKeysLinksResponse struct {
	Self string `json:"self"`
}
