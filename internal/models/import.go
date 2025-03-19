package models

type ImportProviderData struct {
	Owner        string `json:"owner"`
	Repository   string `json:"repository"`
	Token        string `json:"github_token"`
	Tag          string `json:"tag"`
	Name         string `json:"name"`
	GPGPublicKey string `json:"gpg_public_key"`
}

type ImportModuleData struct {
	Owner      string `json:"owner"`
	Repository string `json:"repository"`
	Token      string `json:"github_token"`
	Tag        string `json:"tag"`
	Name       string `json:"name"`
}
