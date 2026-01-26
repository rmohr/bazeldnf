package reducer

import (
	"encoding/xml"
	"os"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/repo"
)

type ReducerPackageLoader interface {
	Load() (*packageInfo, error)
}

// packageInfo captures the package information required by the reducer
type packageInfo struct {
	// the list of available packages
	packages []api.Package

	// mapping of provisions to a list of associated packages
	provides map[string][]*api.Package
}

type RepoLoader struct {
	repoFiles     []string
	architectures []string
	repos         *bazeldnf.Repositories
	cacheHelper   repo.RepoCache
}

func (r RepoLoader) Load() (*packageInfo, error) {
	packageInfo := &packageInfo{
		packages: []api.Package{},
		provides: map[string][]*api.Package{},
	}

	for _, rpmrepo := range r.repoFiles {
		repoFile := &api.Repository{}
		f, err := os.Open(rpmrepo)
		if err != nil {
			return packageInfo, err
		}
		defer f.Close()
		err = xml.NewDecoder(f).Decode(repoFile)
		if err != nil {
			return packageInfo, err
		}
		for i, p := range repoFile.Packages {
			if skip(p.Arch, r.architectures) {
				continue
			}
			packageInfo.packages = append(packageInfo.packages, repoFile.Packages[i])
		}
	}

	cachedRepos, err := r.cacheHelper.CurrentPrimaries(r.repos, r.architectures)
	if err != nil {
		return packageInfo, err
	}
	for _, rpmrepo := range cachedRepos {
		for i, p := range rpmrepo.Packages {
			if skip(p.Arch, r.architectures) {
				continue
			}
			packageInfo.packages = append(packageInfo.packages, rpmrepo.Packages[i])
		}
	}

	for i, _ := range packageInfo.packages {
		FixPackages(&packageInfo.packages[i])
	}

	for i, p := range packageInfo.packages {
		requires := []api.Entry{}
		for _, requirement := range p.Format.Requires.Entries {
			if !strings.HasPrefix(requirement.Name, "(") {
				requires = append(requires, requirement)
			}
		}
		packageInfo.packages[i].Format.Requires.Entries = requires

		for _, p := range p.Format.Provides.Entries {
			packageInfo.provides[p.Name] = append(packageInfo.provides[p.Name], &packageInfo.packages[i])
		}
		for _, file := range p.Format.Files {
			packageInfo.provides[file.Text] = append(packageInfo.provides[file.Text], &packageInfo.packages[i])
		}
	}

	return packageInfo, nil
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
