package models

type ModulesRequest struct {
	Data ModulesDataRequest `json:"data"`
}

type ModulesDataRequest struct {
	Type       string                   `json:"type"`
	Attributes ModulesAttributesRequest `json:"attributes"`
}

type ModulesAttributesRequest struct {
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	Namespace    string `json:"namespace"`
	RegistryName string `json:"registry-name"`
	NoCode       bool   `json:"no-code"`
}
