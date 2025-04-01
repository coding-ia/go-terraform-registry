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

type APIParameters struct {
	Organization string
	Registry     string
	Namespace    string
	Name         string
	Version      string
}
