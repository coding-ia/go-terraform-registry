package types

type ProviderPackageParameters struct {
	Namespace    string
	Name         string
	Version      string
	OS           string
	Architecture string
}

type ProviderVersionParameters struct {
	Namespace string
	Name      string
}

type ModuleVersionParameters struct {
	Namespace string
	Name      string
	System    string
}

type ModuleDownloadParameters struct {
	Namespace string
	Name      string
	System    string
	Version   string
}

type ProviderImport struct {
	DownloadUrl    string
	Filename       string
	GPGASCIIArmor  string
	GPGFingerprint string
	Name           string
	Release        string
	SHASUM         string
	SHASUMUrl      string
	SHASUMSigUrl   string
	Protocols      []string
}
