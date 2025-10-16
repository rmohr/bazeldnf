package reducer

import (
	"fmt"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/sirupsen/logrus"
)

type RepoReducer struct {
	packageInfo      *packageInfo
	implicitRequires []string
	loader           ReducerPackageLoader
}

func (r *RepoReducer) Load() error {
	packageInfo, err := r.loader.Load()
	if err != nil {
		return err
	}
	r.packageInfo = packageInfo
	return nil
}

func (r *RepoReducer) PackageCount() int {
	return len(r.packageInfo.packages)
}

// Checks if user-provided string requesting top-level package matches given package.
// There are various possible matching methods:
// - <package name>
// - <package name>-<version> (version could be also any, possibly empty prefix of package's version)
// - <package name>.<arch>
// - <package name>-<version>.<arch> (needs full version)
func packageMatchesString(pkg *api.Package, req string) bool {
	return req == pkg.Name ||
		strings.HasPrefix(fmt.Sprintf("%s-%s", pkg.Name, pkg.Version.String()), req) && len(req) > len(pkg.Name) ||
		req == fmt.Sprintf("%s.%s", pkg.Name, pkg.Arch) ||
		req == fmt.Sprintf("%s.%s-%s", pkg.Name, pkg.Arch, pkg.Version.String())
}

func (r *RepoReducer) Resolve(packages []string, ignoreMissing bool) (matched []string, involved []*api.Package, err error) {
	packages = append(packages, r.implicitRequires...)
	discovered := map[api.PackageKey]*api.Package{}
	pinned := map[string]*api.Package{}
	for _, req := range packages {
		found := false
		name := ""
		var candidates []*api.Package
		for i, p := range r.packageInfo.packages {
			if packageMatchesString(&p, req) {
				if !found || len(p.Name) < len(name) {
					candidates = []*api.Package{&r.packageInfo.packages[i]}
					name = p.Name
					found = true
				} else if p.Name == name {
					candidates = append(candidates, &r.packageInfo.packages[i])
				}
			}
		}
		if !found && !ignoreMissing {
			return nil, nil, fmt.Errorf("Package %s does not exist", req)
		}

		for i, p := range candidates {
			if selected, ok := discovered[p.Key()]; !ok {
				discovered[p.Key()] = candidates[i]
			} else {
				if selected.Repository.Priority > p.Repository.Priority {
					discovered[p.Key()] = candidates[i]
				}
			}
		}

		if len(candidates) > 0 {
			matched = append(matched, candidates[0].Name)
		}
	}

	for _, v := range discovered {
		pinned[v.Name] = v
	}

	for {
		current := []api.PackageKey{}
		for k := range discovered {
			current = append(current, k)
		}
		for _, p := range current {
			for _, newFound := range r.requires(discovered[p]) {
				if _, exists := discovered[newFound.Key()]; !exists {
					if _, exists := pinned[newFound.Name]; !exists {
						discovered[newFound.Key()] = newFound
					} else {
						logrus.Debugf("excluding %s because of pinned dependency %s", newFound.String(), pinned[newFound.Name].String())
					}
				}
			}
		}
		if len(current) == len(discovered) {
			break
		}
	}

	required := map[string]struct{}{}
	for i, pkg := range discovered {
		for _, req := range pkg.Format.Requires.Entries {
			required[req.Name] = struct{}{}
		}
		involved = append(involved, discovered[i])
	}
	// remove all provides which are not required in the reduced set
	for i, pkg := range involved {
		provides := []api.Entry{}
		for j, prov := range pkg.Format.Provides.Entries {
			if _, exists := required[prov.Name]; exists || prov.Name == pkg.Name {
				provides = append(provides, pkg.Format.Provides.Entries[j])
			}
		}
		involved[i].Format.Provides.Entries = provides
	}

	return matched, involved, nil
}

func (r *RepoReducer) requires(p *api.Package) (wants []*api.Package) {
	for _, requires := range p.Format.Requires.Entries {
		if val, exists := r.packageInfo.provides[requires.Name]; exists {
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

func NewRepoReducer(repos *bazeldnf.Repositories, repoFiles []string, baseSystem string, architectures []string, cacheHelper *repo.CacheHelper) *RepoReducer {
	implicitRequires := make([]string, 0, 1)
	if baseSystem != "" {
		implicitRequires = append(implicitRequires, baseSystem)
	}
	return &RepoReducer{
		packageInfo:      nil,
		implicitRequires: implicitRequires,
		loader: RepoLoader{
			repoFiles:     repoFiles,
			architectures: architectures,
			repos:         repos,
			cacheHelper:   cacheHelper,
		},
	}
}

func Resolve(repos *bazeldnf.Repositories, repoFiles []string, baseSystem string, architectures []string, packages []string, ignoreMissing bool) (matched []string, involved []*api.Package, err error) {
	repoReducer := NewRepoReducer(repos, repoFiles, baseSystem, architectures, repo.NewCacheHelper())
	logrus.Info("Loading packages.")
	if err := repoReducer.Load(); err != nil {
		return nil, nil, err
	}
	logrus.Infof("loaded %d packages", repoReducer.PackageCount())
	logrus.Info("Initial reduction of involved packages.")
	return repoReducer.Resolve(packages, ignoreMissing)
}
