package main

import (
	"cmp"
	"encoding/json"
	"os"
	"regexp"

	"slices"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/rmohr/bazeldnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type BzlmodOpts struct {
	out       string
	repoFiles []string
}

var bzlmodopts = BzlmodOpts{}

func NewBzlmodCmd() *cobra.Command {

	bzlmodCmd := &cobra.Command{
		Use:   "bzlmod",
		Short: "Manage bazeldnf bzlmod lock file",
		Long:  `From a set of dependencies keeps the bazeldnf bzlmod json lock file up to date`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bzlmodopts.RunE(cmd, args)
		},
	}

	addResolveHelperFlags(bzlmodCmd)
	repo.AddCacheHelperFlags(bzlmodCmd)

	bzlmodCmd.Flags().StringVarP(&bzlmodopts.out, "output", "o", "/dev/stdout", "Output file for the lock contents (defaults to /dev/stdout)")
	bzlmodCmd.Flags().StringArrayVarP(&bzlmodopts.repoFiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	bzlmodCmd.Args = cobra.MinimumNArgs(1)

	return bzlmodCmd
}

type ResolvedResult struct {
	Install      []*api.Package `json:"install"`
	ForceIgnored []*api.Package `json:"force_ignored"`
}

type InstalledPackage struct {
	Name         string   `json:"name"`
	Sha256       string   `json:"sha256"`
	Href         string   `json:"href"`
	Repository   string   `json:"repository"`
	Dependencies []string `json:"dependencies"`
}

func (i *InstalledPackage) setDependencies(pkgs []string) {
	i.Dependencies = make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg == i.Name {
			logrus.Infof("Ignoring self-dependency %s", pkg)
			continue
		}
		i.Dependencies = append(i.Dependencies, pkg)
	}
}

type BzlmodLockFile struct {
	CommandLineArguments []string            `json:"cli-arguments,omitempty"`
	Repositories         map[string][]string `json:"repositories"`
	Packages             []InstalledPackage  `json:"packages"`
	Targets              []string            `json:"targets,omitempty"`
	ForceIgnored         []string            `json:"ignored,omitempty"`
}

func DumpJSON(result ResolvedResult, targets []string, cmdline []string) ([]byte, error) {
	forceIgnored := make(map[string]bool)
	allPackages := make(map[string]InstalledPackage)
	providers := make(map[string]string)
	repositories := make(map[string]*bazeldnf.Repository)

	for _, forceIgnoredPackage := range result.ForceIgnored {
		forceIgnored[forceIgnoredPackage.Name] = true

		for _, entry := range forceIgnoredPackage.Format.Provides.Entries {
			providers[entry.Name] = forceIgnoredPackage.Name
		}

		for _, entry := range forceIgnoredPackage.Format.Files {
			providers[entry.Text] = forceIgnoredPackage.Name
		}
	}

	for _, installPackage := range result.Install {
		deps := make([]string, 0, len(installPackage.Format.Requires.Entries))

		for _, entry := range installPackage.Format.Requires.Entries {
			deps = append(deps, entry.Name)
		}

		for _, entry := range installPackage.Format.Provides.Entries {
			providers[entry.Name] = installPackage.Name
		}

		for _, entry := range installPackage.Format.Files {
			providers[entry.Text] = installPackage.Name
		}

		slices.Sort(deps)
		repositories[installPackage.Repository.Name] = installPackage.Repository

		allPackages[installPackage.Name] = InstalledPackage{
			Name:         installPackage.Name,
			Sha256:       installPackage.Checksum.Text,
			Href:         installPackage.Location.Href,
			Repository:   installPackage.Repository.Name,
			Dependencies: deps,
		}
	}

	alreadyInstalled := make(map[string]bool)
	for _, name := range sortedKeys(allPackages) {
		if _, ignored := forceIgnored[name]; ignored {
			continue
		}

		requires := allPackages[name].Dependencies
		deps, err := computeDependencies(requires, providers, forceIgnored)
		if err != nil {
			return nil, err
		}
		alreadyInstalled[name] = true

		// RPMs may have circular dependencies, even depend on themselves.
		// we need to ignore such dependency
		non_cyclic_deps := make([]string, 0, len(deps))
		for _, dep := range deps {
			if alreadyInstalled[dep] {
				continue
			}
			non_cyclic_deps = append(non_cyclic_deps, dep)
		}
		entry := allPackages[name]
		entry.setDependencies(non_cyclic_deps)
		allPackages[name] = entry
	}

	allPackages, forceIgnored = garbageCollect(targets, forceIgnored, allPackages)

	packageNames := sortedKeys(allPackages)

	sortedPackages := make([]InstalledPackage, 0, len(packageNames))
	for _, name := range packageNames {
		sortedPackages = append(sortedPackages, allPackages[name])
	}

	ignoredPackages := sortedKeys(forceIgnored)

	slices.Sort(ignoredPackages)

	lockFile := BzlmodLockFile{
		CommandLineArguments: cmdline,
		ForceIgnored:         ignoredPackages,
		Packages:             sortedPackages,
		Repositories:         make(map[string][]string),
	}

	for mirrorName, repository := range repositories {
		lockFile.Repositories[mirrorName] = repository.Mirrors
	}

	if len(targets) > 0 {
		lockFile.Targets = targets
	}

	return json.MarshalIndent(lockFile, "", "\t")
}

func computeDependencies(requires []string, providers map[string]string, ignored map[string]bool) ([]string, error) {
	deps := make(map[string]bool)
	for _, req := range requires {
		if ignored[req] {
			logrus.Debugf("Ignoring dependency %s", req)
			continue
		}
		logrus.Debugf("Resolving dependency %s", req)
		provider, ok := providers[req]
		if !ok {
			logrus.Warnf("could not find provider for %s", req)
			continue
		}
		logrus.Debugf("Found provider %s for %s", provider, req)
		if ignored[provider] {
			logrus.Debugf("Ignoring provider %s for %s", provider, req)
			continue
		}
		deps[provider] = true
	}
	return sortedKeys(deps), nil
}

func shouldInclude(target string, ignoreRegex []string) bool {
	for _, rex := range ignoreRegex {
		if match, err := regexp.MatchString(rex, target); err != nil {
			logrus.Errorf("Failed to match package %s with regex '%v': %v", target, rex, err)
			return false
		} else if match {
			logrus.Debugf("will not include %s as it matched %s", target, rex)
			return false
		}
	}
	logrus.Debugf("including %s", target)
	return true
}

func exploreAllDependencies(input *api.Package, allPackages map[string]*api.Package, previouslyExplored map[string]bool) []*api.Package {
	alreadyExplored := make(map[string]*api.Package, 0)
	pending := []*api.Package{input}

	for len(pending) > 0 {
		current := pending[0]
		pending = pending[1:]

		if _, explored := previouslyExplored[current.Name]; explored {
			continue
		}

		if _, explored := alreadyExplored[current.Name]; explored {
			continue
		}

		alreadyExplored[current.Name] = current

		for _, entry := range current.Format.Requires.Entries {
			t, ok := allPackages[entry.Name]
			if !ok {
				continue
			}
			pending = append(pending, t)
		}
	}

	output := maps.Values(alreadyExplored)
	slices.SortStableFunc(output, func(a, b *api.Package) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return output
}

func filterIgnores(rpmsRequested []string, allAvailable []*api.Package, ignoreRegex []string) ([]*api.Package, []*api.Package) {
	/*
	 * Given a list of rpmsRequested and a list of ignoreRegex then go through all the resolved rpms
	 * in allAvailable to install and filter only those needed.
	 */
	if len(ignoreRegex) == 0 {
		return allAvailable, []*api.Package{}
	}

	allAvailablePerName := make(map[string]*api.Package, 0)

	for _, rpm := range allAvailable {
		allAvailablePerName[rpm.Name] = rpm

		for _, entry := range rpm.Format.Provides.Entries {
			allAvailablePerName[entry.Name] = rpm
		}

		for _, entry := range rpm.Format.Files {
			allAvailablePerName[entry.Text] = rpm
		}
	}

	toInstall := make([]*api.Package, 0)
	ignored := make(map[string]*api.Package, 0)

	requested := map[string]bool{}
	for _, rpm := range rpmsRequested {
		requested[rpm] = true
	}

	explored := make(map[string]bool, 0)
	for _, rpm := range rpmsRequested {
		target, ok := allAvailablePerName[rpm]
		if !ok {
			logrus.Errorf("failed to match %s", rpm)
			continue
		}

		pending := []*api.Package{target}
		for len(pending) > 0 {
			current := pending[0]
			pending = pending[1:]
			if _, alreadyExplored := explored[current.Name]; alreadyExplored {
				continue
			}

			logrus.Debugf("processing %s", current.Name)
			if !shouldInclude(current.Name, ignoreRegex) {
				toIgnore := exploreAllDependencies(current, allAvailablePerName, explored)
				for _, d := range toIgnore {
					// don't exclude those things that were explicitly requested
					if _, isRequested := requested[d.Name]; isRequested {
						continue
					}
					logrus.Debugf("excluding %s", d.Name)
					ignored[d.Name] = d
					explored[d.Name] = true
				}
				continue
			}

			explored[current.Name] = true

			toInstall = append(toInstall, current)
			for _, dep := range current.Format.Requires.Entries {
				p, ok := allAvailablePerName[dep.Name]
				if ok {
					pending = append(pending, p)
				}
			}
		}
	}

	ignoredPackages := maps.Values(ignored)
	slices.SortStableFunc(ignoredPackages, func(a, b *api.Package) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return toInstall, ignoredPackages
}

func garbageCollect(targets []string, forceIgnored map[string]bool, packages map[string]InstalledPackage) (map[string]InstalledPackage, map[string]bool) {
	reachedPackages := map[string]bool{}

	for _, target := range targets {
		reachedPackages[target] = true
		for _, dep := range packages[target].Dependencies {
			reachedPackages[dep] = true
		}
	}

	for _, pkg := range packages {
		if _, isReached := reachedPackages[pkg.Name]; !isReached {
			forceIgnored[pkg.Name] = true
		}
	}

	for pkg, _ := range forceIgnored {
		delete(packages, pkg)
	}

	return packages, forceIgnored
}

func packageNames(packages []*api.Package) []string {
	output := make([]string, 0)
	for _, p := range packages {
		output = append(output, p.Name)
	}
	slices.Sort(output)
	return output
}

func (opts *BzlmodOpts) RunE(cmd *cobra.Command, rpms []string) error {
	logrus.Info("Loading repo files")
	repos, err := repo.LoadRepoFiles(bzlmodopts.repoFiles)
	if err != nil {
		return err
	}

	logrus.Debugf("loaded repo files: %+v", repos)

	repoReducer := reducer.NewRepoReducer(repos, []string{}, "", resolvehelperopts.arch, repo.NewCacheHelper())

	logrus.Info("Loading packages.")
	if err := repoReducer.Load(); err != nil {
		return err
	}

	logrus.Infof("Initial reduction to resolve dependencies for targets %v", rpms)
	matched, involved, err := repoReducer.Resolve(rpms, resolvehelperopts.ignoreMissing)
	if err != nil {
		return err
	}

	solver := sat.NewResolver(resolvehelperopts.nobest)
	logrus.Infof("Loading involved packages into the rpmtreer: %d", len(involved))
	err = solver.LoadInvolvedPackages(involved, []string{}, resolvehelperopts.onlyAllowRegex)
	if err != nil {
		return err
	}

	logrus.Infof("Adding required packages to the rpmtreer: %d", len(matched))
	err = solver.ConstructRequirements(matched)
	if err != nil {
		return err
	}

	logrus.Info("Solving.")
	install, _, _, err := solver.Resolve()
	if err != nil {
		return err
	}

	logrus.Debugf("resolver install: %v", install)

	actualInstall, forceIgnored := filterIgnores(rpms, install, resolvehelperopts.forceIgnoreRegex)
	logrus.Debugf("before GC actual install(%d): %+v", len(actualInstall), packageNames(actualInstall))
	logrus.Debugf("before GC actual ignored(%d): %+v", len(forceIgnored), packageNames(forceIgnored))

	result := ResolvedResult{Install: actualInstall, ForceIgnored: forceIgnored}

	data, err := DumpJSON(result, rpms, os.Args[2:])

	if err != nil {
		return err
	}

	logrus.Info("Writing lock file.")

	return os.WriteFile(opts.out, data, 0644)
}
