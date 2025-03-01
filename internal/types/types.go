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
