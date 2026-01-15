package bazeldnf

type RPM struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Arch         string   `json:"arch"`
	Integrity    string   `json:"integrity"`
	URLs         []string `json:"urls"`
	Repository   string   `json:"repository"`
	Dependencies []string `json:"dependencies"`
}

type Config struct {
	CommandLineArguments []string            `json:"cli-arguments,omitempty"`
	Name                 string              `json:"name"`
	Repositories         map[string][]string `json:"repositories"`
	RPMs                 []*RPM              `json:"rpms"`
	Targets              []string            `json:"targets,omitempty"`
	ForceIgnored         []string            `json:"ignored,omitempty"`
}
