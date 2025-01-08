package reducer

import (
	"encoding/xml"
	"os"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

type RepoCache interface {
	CurrentPrimaries(repos *bazeldnf.Repositories, arch string) (primaries []*api.Repository, err error)
}

type ReducerPackageLoader interface {
	Load() ([]api.Package, error)
}

type RepoLoader struct {
	repoFiles     []string
	arch          string
	architectures []string
	repos         *bazeldnf.Repositories
	cacheHelper   RepoCache
}

func (r RepoLoader) Load() ([]api.Package, error) {
	packages := []api.Package{}

	for _, rpmrepo := range r.repoFiles {
		repoFile := &api.Repository{}
		f, err := os.Open(rpmrepo)
		if err != nil {
			return packages, err
		}
		defer f.Close()
		err = xml.NewDecoder(f).Decode(repoFile)
		if err != nil {
			return packages, err
		}
		for i, p := range repoFile.Packages {
			if skip(p.Arch, r.architectures) {
				continue
			}
			packages = append(packages, repoFile.Packages[i])
		}
	}

	cachedRepos, err := r.cacheHelper.CurrentPrimaries(r.repos, r.arch)
	if err != nil {
		return packages, err
	}
	for _, rpmrepo := range cachedRepos {
		for i, p := range rpmrepo.Packages {
			if skip(p.Arch, r.architectures) {
				continue
			}
			packages = append(packages, rpmrepo.Packages[i])
		}
	}

	for i, _ := range packages {
		FixPackages(&packages[i])
	}

	return packages, nil
}

// FixPackages contains hacks which should probably not have to exist
func FixPackages(p *api.Package) {
	// FIXME: This is not a proper modules support for python. We should properly resolve `alternative(python)` and
	// not have to add such a hack. On the other hand this seems to have been reverted in fedora and only exists in centos stream.
	if p.Name == "platform-python" {
		p.Format.Provides.Entries = append(p.Format.Provides.Entries, api.Entry{
			Name: "/usr/libexec/platform-python",
		})
		var requires []api.Entry
		for _, entry := range p.Format.Requires.Entries {
			if entry.Name != "/usr/libexec/platform-python" {
				requires = append(requires, entry)
			}
		}
		p.Format.Requires.Entries = requires
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
