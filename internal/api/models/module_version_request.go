package models

type ModuleVersionsRequest struct {
	Data ModuleVersionsDataRequest `json:"data"`
}

type ModuleVersionsDataRequest struct {
	Type       string                          `json:"type"`
	Attributes ModuleVersionsAttributesRequest `json:"attributes"`
}

type ModuleVersionsAttributesRequest struct {
	Version   string `json:"version"`
	CommitSHA string `json:"commit-sha"`
}
