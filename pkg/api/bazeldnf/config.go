package bazeldnf

type RPM struct {
	Name string `json:"name"`
	SHA256 string `json:"sha256"`
	URLs []string `json:"urls"`
}

type Config struct {
	Name string `json:"name"`
	RPMs []RPM `json:"rpms"`
}
