package models

type TerraformAvailableModule struct {
	Modules []TerraformAvailableModuleVersions `json:"modules"`
}

type TerraformAvailableModuleVersions struct {
	Versions []TerraformAvailableModuleVersion `json:"versions"`
}

type TerraformAvailableModuleVersion struct {
	Version string `json:"version"`
}
