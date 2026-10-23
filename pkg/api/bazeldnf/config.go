package bazeldnf

type RPM struct {
	Id           string   `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	SHA256       string   `json:"sha256,omitempty"`
	Integrity    string   `json:"integrity,omitempty"`
	URLs         []string `json:"urls"`
	Repository   string   `json:"repository,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type Config struct {
	CommandLineArguments []string            `json:"cli-arguments,omitempty"`
	Name                 string              `json:"name,omitempty"`
	Repositories         map[string][]string `json:"repositories,omitempty"`
	RPMs                 []*RPM              `json:"rpms"`
	Targets              []string            `json:"targets,omitempty"`
	ForceIgnored         []string            `json:"ignored,omitempty"`
}
