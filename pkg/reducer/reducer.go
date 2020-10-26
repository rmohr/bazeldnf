package reducer

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rmohr/bazel-dnf/pkg/api"
	"github.com/sirupsen/logrus"
)

type RepoReducer struct {
	packages         []api.Package
	lang             string
	repoFiles        []string
	provides         map[string][]*api.Package
	implicitRequires []string
	arch             string
	architectures    []string
}

func (r *RepoReducer) Load() error {
	for _, repo := range r.repoFiles {
		repoFile := &api.Repository{}
		f, err := os.Open(repo)
		if err != nil {
			return err
		}
		defer f.Close()
		err = xml.NewDecoder(f).Decode(repoFile)
		if err != nil {
			return err
		}
		r.packages = append(r.packages, repoFile.Packages...)
	}

	for i, p := range r.packages {
		canBeInterpreted := true
		for _, requires := range p.Format.Requires.Entries {
			if strings.HasPrefix(requires.Name, "(") {
				canBeInterpreted = false
			}
		}
		if skip(p.Arch, r.architectures) || !canBeInterpreted {
			continue
		}

		for _, provides := range p.Format.Provides.Entries {
			r.provides[provides.Name] = append(r.provides[provides.Name], &r.packages[i])
		}
		for _, file := range p.Format.Files {
			r.provides[file.Text] = append(r.provides[file.Text], &r.packages[i])
		}
	}
	return nil
}

func (r *RepoReducer) Resolve(packages []string) (involved []*api.Package, err error) {
	packages = append(packages, r.implicitRequires...)
	var wants []*api.Package
	discovered := map[string]*api.Package{}
	for _, req := range packages {
		for _, p := range r.packages {
			if p.Name == req {
				wants = append(wants, &p)
				discovered[p.Name] = &p
				break
			}
		}
	}
	if len(wants) == 0 {
		return nil, fmt.Errorf("Package %s does not exist", packages[0])
	}

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
	testrepo := &api.Repository{}
	for _, pkg := range involved {
		testrepo.Packages = append(testrepo.Packages, *pkg)
	}
	data, _ := xml.MarshalIndent(testrepo, "", "  ")
	ioutil.WriteFile("test.xml", data, 0666)
	return involved, nil
}

func (r *RepoReducer) requires(p *api.Package) (wants []*api.Package) {
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

func NewRepoReducer(repoFiles []string, lang string, fedoraRelease string, arch string) *RepoReducer {
	return &RepoReducer{
		packages:         nil,
		lang:             lang,
		implicitRequires: []string{fedoraRelease},
		repoFiles:        repoFiles,
		provides:         map[string][]*api.Package{},
		architectures:    []string{"noarch", arch},
	}
}

func skip(arch string, arches []string) bool {
	skip := true
	for _, a := range arches {
		if a == arch {
			skip = false
			break
		}
	}
	return skip
}
