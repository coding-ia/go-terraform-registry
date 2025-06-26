package models

type GPGKeysListResponse struct {
	Data  []GPGKeysDataResponse `json:"data"`
	Links Links                 `json:"links"`
	Meta  Meta                  `json:"meta"`
}
