package models

type ModuleVersionsResponse struct {
	Data ModuleVersionsDataResponse `json:"data"`
}

type ModuleVersionsDataResponse struct {
	ID            string                              `json:"id"`
	Type          string                              `json:"type"`
	Attributes    ModuleVersionsAttributesResponse    `json:"attributes"`
	Relationships ModuleVersionsRelationshipsResponse `json:"relationships"`
	Links         ModuleVersionsLinksResponse         `json:"links"`
}

type ModuleVersionsAttributesResponse struct {
	Source    string `json:"source"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	CreatedAt string `json:"created-at"`
	UpdatedAt string `json:"updated-at"`
}

type ModuleVersionsRelationshipsResponse struct {
	RegistryModule ModuleVersionsRegistryModuleResponse `json:"registry-module"`
}

type ModuleVersionsRegistryModuleResponse struct {
	Data ModuleVersionsRegistryModuleDataResponse `json:"data"`
}

type ModuleVersionsRegistryModuleDataResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ModuleVersionsLinksResponse struct {
	Upload string `json:"upload"`
}
