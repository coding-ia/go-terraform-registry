package models

type GPGKeysRequest struct {
	Data GPGKeysDataRequest `json:"data"`
}

type GPGKeysDataRequest struct {
	Type       string                   `json:"type"`
	Attributes GPGKeysAttributesRequest `json:"attributes"`
}

type GPGKeysAttributesRequest struct {
	Namespace  string `json:"namespace"`
	AsciiArmor string `json:"ascii_armor"`
}
