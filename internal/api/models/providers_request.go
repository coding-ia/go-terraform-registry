package models

type ProvidersRequest struct {
	Data ProvidersDataRequest `json:"data"`
}

type ProvidersDataRequest struct {
	Type       string                     `json:"type"`
	Attributes ProvidersAttributesRequest `json:"attributes"`
}

type ProvidersAttributesRequest struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RegistryName string `json:"registry-name"`
}
