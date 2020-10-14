package repoquery

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/rmohr/bazel-dnf/pkg/api"
	"github.com/sirupsen/logrus"
)

type RepoQuery struct {
	repo      *api.Repository
	lang string
	repoFiles []string
	provides  map[string][]*api.Package
}

func (r *RepoQuery) Load() error {
	for _, repo := range r.repoFiles {
		f, err := os.Open(repo)
		if err != nil {
			return err
		}
		r.repo = &api.Repository{}
		err = xml.NewDecoder(f).Decode(r.repo)
		if err != nil {
			return err
		}

		for i, p := range r.repo.Packages {
			if p.Arch == "i686" {
				continue
			}
			for _, provides := range p.Format.Provides.Entries {
				r.provides[provides.Name] = append(r.provides[provides.Name], &r.repo.Packages[i])
			}
			for _, file := range p.Format.Files {
				r.provides[file.Text] = append(r.provides[file.Text], &r.repo.Packages[i])
			}
		}
	}
	return nil
}

func (r *RepoQuery) Resolve(packages []string) (involved []*api.Package, err error) {
	var wants []*api.Package
	for _, p := range r.repo.Packages {
		if p.Name == packages[0] {
			wants = append(wants, &p)
			break
		}
	}
	if len(wants) == 0 {
		return nil, fmt.Errorf("Package %s does not exist", packages[0])
	}

	discovered := map[string]*api.Package{}
	discovered[wants[0].Name] = wants[0]

	for {
		current := []string{}
		for k := range discovered {
			current = append(current, k)
		}
		for _, p := range current {
			for _, newFound := range r.requires(discovered[p]) {
				if _, exists := discovered[newFound.Name]; !exists {
					discovered[newFound.Name] = newFound
				}
			}
		}
		if len(current) == len(discovered) {
			break
		}
	}

	current := []string{}
	for k := range discovered {
		involved = append(involved, r.provides[k]...)
		current = append(current, k)
	}
	fmt.Println(strings.Join(current, ","))
	return involved, nil
}

func (r *RepoQuery) requires(p *api.Package) (wants []*api.Package) {
	for _, requires := range p.Format.Requires.Entries {
		if val, exists := r.provides[requires.Name]; exists {

			var packages []string
			for _, p := range val {
				packages = append(packages, p.Name)
			}
			logrus.Debugf("%s wants %v because of %v\n", p.Name, packages, requires)
			wants = append(wants, val...)
		} else {
			logrus.Debugf("%s requires %v which can't be satisfied\n", p.Name, requires)
		}
	}
	return wants
}

func NewRepoQuerier(repoFiles []string, lang string) *RepoQuery {
	return &RepoQuery{
		repo:      nil,
		lang: lang,
		repoFiles: repoFiles,
		provides:  map[string][]*api.Package{},
	}
}
