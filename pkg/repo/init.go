package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"sigs.k8s.io/yaml"
)

type RepoInit struct {
	OS                 string
	Arch               string
	PrimaryMetaLinkURL string
	UpdateMetaLinkURL  string
	RepoFile           string
}

func (r *RepoInit) Init() error {

	_, err := os.Stat(r.RepoFile)
	if !os.IsNotExist(err) {
		return fmt.Errorf("repository file %s already exists.", r.RepoFile)
	}
	repos := &bazeldnf.Repositories{
		Repositories: []bazeldnf.Repository{
			{
				Name:     fmt.Sprintf("%s-%s-primary-repo", r.OS, r.Arch),
				Disabled: false,
				Metalink: r.PrimaryMetaLinkURL,
				Baseurl:  "",
				Arch:     r.Arch,
			},
			{
				Name:     fmt.Sprintf("%s-%s-update-repo", r.OS, r.Arch),
				Disabled: false,
				Metalink: r.UpdateMetaLinkURL,
				Baseurl:  "",
				Arch:     r.Arch,
			},
		},
	}
	data, err := yaml.Marshal(repos)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(r.RepoFile, data, 0660)
}

func NewRemoteInit(os string, arch string, repoFile string) *RepoInit {
	os = strings.TrimPrefix(os, "f")
	return &RepoInit{
		OS:                 os,
		Arch:               arch,
		RepoFile:           repoFile,
		PrimaryMetaLinkURL: fmt.Sprintf("https://mirrors.fedoraproject.org/metalink?repo=fedora-%s&arch=%s", os, arch),
		UpdateMetaLinkURL:  fmt.Sprintf("https://mirrors.fedoraproject.org/metalink?repo=updates-released-f%s&arch=%s", os, arch),
	}
}

func LoadRepoFile(file string) (*bazeldnf.Repositories, error) {
	repofile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	repos := &bazeldnf.Repositories{}
	err = yaml.Unmarshal(repofile, repos)
	if err != nil {
		return nil, err
	}
	return repos, err
}
