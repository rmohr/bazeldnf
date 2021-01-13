package reducer

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/repo"
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
	repos            *bazeldnf.Repositories
	cacheHelper      *repo.CacheHelper
}

func (r *RepoReducer) Load() error {
	for _, rpmrepo := range r.repoFiles {
		repoFile := &api.Repository{}
		f, err := os.Open(rpmrepo)
		if err != nil {
			return err
		}
		defer f.Close()
		err = xml.NewDecoder(f).Decode(repoFile)
		if err != nil {
			return err
		}
		for i, p := range repoFile.Packages {
			if skip(p.Arch, r.architectures) {
				continue
			}
			r.packages = append(r.packages, repoFile.Packages[i])
		}
	}
	repos, err := r.cacheHelper.CurrentPrimaries(r.repos, r.arch)
	if err != nil {
		return err
	}
	for _, rpmrepo := range repos {
		for i, p := range rpmrepo.Packages {
			if skip(p.Arch, r.architectures) {
				continue
			}
			r.packages = append(r.packages, rpmrepo.Packages[i])
		}
	}

	for i, p := range r.packages {
		requires := []api.Entry{}
		for _, requirement := range p.Format.Requires.Entries {
			if !strings.HasPrefix(requirement.Name, "(") {
				requires = append(requires, requirement)
			}
		}
		r.packages[i].Format.Requires.Entries = requires

		for _, provides := range p.Format.Provides.Entries {
			r.provides[provides.Name] = append(r.provides[provides.Name], &r.packages[i])
		}
		for _, file := range p.Format.Files {
			r.provides[file.Text] = append(r.provides[file.Text], &r.packages[i])
		}
	}
	return nil
}

func (r *RepoReducer) Resolve(packages []string) (matched []string, involved []*api.Package, err error) {
	packages = append(packages, r.implicitRequires...)
	discovered := map[string]*api.Package{}
	for _, req := range packages {
		found := false
		name := ""
		var candidates []*api.Package
		for i, p := range r.packages {
			if strings.HasPrefix(p.String(), req) {
				if strings.HasPrefix(req, p.Name) {
					if !found || len(p.Name) < len(name) {
						candidates = []*api.Package{&r.packages[i]}
						name = p.Name
						found = true
					} else if p.Name == name {
						candidates = append(candidates, &r.packages[i])
					}
				}
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("Package %s does not exist", req)
		}
		for i, p := range candidates {
			discovered[p.String()] = candidates[i]
		}

		if len(candidates) > 0 {
			matched = append(matched, candidates[0].Name)
		}
	}

	for {
		current := []string{}
		for k := range discovered {
			current = append(current, k)
		}
		for _, p := range current {
			for _, newFound := range r.requires(discovered[p]) {
				if _, exists := discovered[newFound.String()]; !exists {
					discovered[newFound.String()] = newFound
				}
			}
		}
		if len(current) == len(discovered) {
			break
		}
	}

	for i, _ := range discovered {
		involved = append(involved, discovered[i])
	}
	return matched, involved, nil
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

func NewRepoReducer(repos *bazeldnf.Repositories, repoFiles []string, lang string, fedoraRelease string, arch string, cachDir string) *RepoReducer {
	return &RepoReducer{
		packages:         nil,
		lang:             lang,
		implicitRequires: []string{fedoraRelease},
		repoFiles:        repoFiles,
		provides:         map[string][]*api.Package{},
		architectures:    []string{"noarch", arch},
		arch:             arch,
		repos:            repos,
		cacheHelper:      &repo.CacheHelper{CacheDir: cachDir},
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
