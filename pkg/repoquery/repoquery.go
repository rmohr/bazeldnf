package repoquery

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rmohr/bazel-dnf/pkg/api"
	"github.com/sirupsen/logrus"
)

type RepoQuery struct {
	packages      []api.Package
	lang      string
	repoFiles []string
	provides  map[string][]*api.Package
}

func (r *RepoQuery) Load() error {
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
		if p.Arch == "i686" {
			continue
		}
		// remove langpack references
		newRequires :=  []api.Entry{}
		for _, requires := range p.Format.Requires.Entries {
			if strings.HasPrefix(requires.Name, "glibc-langpack") {
				requires.Name="glibc-langpack-en"
			}
			newRequires=append(newRequires, requires)
		}
		r.packages[i].Format.Requires.Entries = newRequires
		for _, provides := range p.Format.Provides.Entries {
			r.provides[provides.Name] = append(r.provides[provides.Name], &r.packages[i])
		}
		for _, file := range p.Format.Files {
			r.provides[file.Text] = append(r.provides[file.Text], &r.packages[i])
		}
	}
	return nil
}

func (r *RepoQuery) Resolve(packages []string) (involved []*api.Package, err error) {
	fmt.Println(len(r.packages))
	var wants []*api.Package
	discovered := map[string]*api.Package{}
	for _, req := range packages {
		for _, p := range r.packages {
			if p.Name == req {
				if p.Arch == "i686" {
					continue
				}
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
		testrepo.Packages=append(testrepo.Packages, *pkg)
	}
	data, _ := xml.MarshalIndent(testrepo, "", "  ")
	ioutil.WriteFile("test.xml", data, 0666)
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
		packages:      nil,
		lang:      lang,
		repoFiles: repoFiles,
		provides:  map[string][]*api.Package{},
	}
}
