package bazeldnf

type Repositories struct {
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name string `json:"name"`
	Disabled bool `json:"disabled,omitempty"`
	Metalink string `json:"metalink,omitempty"`
	Baseurl string `json:"baseurl,omitempty"`
	Arch string `json:"arch"`
	Mirrors []string `json:"mirrors,omitempty"`
}