package models

type ProvidersResponse struct {
	Data ProvidersDataResponse `json:"data"`
}

type ProvidersDataResponse struct {
	ID            string                         `json:"id"`
	Type          string                         `json:"type"`
	Attributes    ProvidersAttributesResponse    `json:"attributes"`
	Relationships ProvidersRelationshipsResponse `json:"relationships"`
	Links         ProvidersSelfLink              `json:"links"`
}

type ProvidersAttributesResponse struct {
	Name         string                       `json:"name"`
	Namespace    string                       `json:"namespace"`
	RegistryName string                       `json:"registry-name"`
	CreatedAt    string                       `json:"created-at"`
	UpdatedAt    string                       `json:"updated-at"`
	Permissions  ProvidersPermissionsResponse `json:"permissions"`
}

type ProvidersPermissionsResponse struct {
	CanDelete bool `json:"can-delete"`
}

type ProvidersRelationshipsResponse struct {
	Organization ProvidersRelationshipSingleResponse `json:"organization"`
	Versions     ProvidersRelationshipListResponse   `json:"versions"`
}

type ProvidersRelationshipSingleResponse struct {
	Data ProvidersRelationshipDataResponse `json:"data"`
}

type ProvidersRelationshipListResponse struct {
	Data  []ProvidersRelationshipDataResponse `json:"data"`
	Links ProvidersRelatedLinkResponse        `json:"links"`
}

type ProvidersRelationshipDataResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ProvidersRelatedLinkResponse struct {
	Related string `json:"related"`
}

type ProvidersSelfLink struct {
	Self string `json:"self"`
}
