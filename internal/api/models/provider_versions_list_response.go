package models

type ProviderVersionsListResponse struct {
	Data  []ProviderVersionsDataResponse `json:"data"`
	Links Links                          `json:"links"`
	Meta  Meta                           `json:"meta"`
}
