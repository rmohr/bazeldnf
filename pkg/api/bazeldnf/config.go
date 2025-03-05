package bazeldnf

type RPM struct {
	Name         string   `json:"name"`
	Integrity    string   `json:"integrity"`
	URLs         []string `json:"urls"`
	Repository   string   `json:"repository"`
	Dependencies []string `json:"dependencies"`
}

func (i *RPM) SetDependencies(pkgs []string) {
	i.Dependencies = make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg == i.Name {
			continue
		}
		i.Dependencies = append(i.Dependencies, pkg)
	}
}

type Config struct {
	CommandLineArguments []string            `json:"cli-arguments,omitempty"`
	Name                 string              `json:"name"`
	Repositories         map[string][]string `json:"repositories"`
	RPMs                 []*RPM              `json:"rpms"`
	Targets              []string            `json:"targets,omitempty"`
	ForceIgnored         []string            `json:"ignored,omitempty"`
}
