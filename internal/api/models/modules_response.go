package models

type ModulesResponse struct {
	Data ModulesDataResponse `json:"data"`
}

type ModulesDataResponse struct {
	ID            string                       `json:"id"`
	Type          string                       `json:"type"`
	Attributes    ModulesAttributesResponse    `json:"attributes"`
	Relationships ModulesRelationshipsResponse `json:"relationships"`
	Links         ModulesLinksResponse         `json:"links"`
}

type ModulesAttributesResponse struct {
	Name            string                     `json:"name"`
	Namespace       string                     `json:"namespace"`
	RegistryName    string                     `json:"registry-name"`
	Provider        string                     `json:"provider"`
	Status          string                     `json:"status"`
	VersionStatuses []string                   `json:"version-statuses"`
	CreatedAt       string                     `json:"created-at"`
	UpdatedAt       string                     `json:"updated-at"`
	Permissions     ModulesPermissionsResponse `json:"permissions"`
}

type ModulesPermissionsResponse struct {
	CanDelete bool `json:"can-delete"`
	CanResync bool `json:"can-resync"`
	CanRetry  bool `json:"can-retry"`
}

type ModulesRelationshipsResponse struct {
	Organization ModulesOrganizationResponse `json:"organization"`
}

type ModulesOrganizationResponse struct {
	Data ModulesOrganizationDataResponse `json:"data"`
}

type ModulesOrganizationDataResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ModulesLinksResponse struct {
	Self string `json:"self"`
}
