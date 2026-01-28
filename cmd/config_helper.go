package main

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

// makeId creates an opaque, deterministic string identifier, unique for each package present in the config.
func makeId(pkg *api.Package) string {
	id := pkg.Name
	if pkg.Arch != "" {
		id += "." + pkg.Arch
	}
	return id
}

func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}

func sortedPackages(pkgs []*api.Package) []*api.Package {
	return slices.SortedFunc(slices.Values(pkgs), func(p1, p2 *api.Package) int {
		return cmp.Or(
			cmp.Compare(p1.Name, p2.Name),
		)
	})
}

func toConfig(install, forceIgnored []*api.Package, targets []string, cmdline []string) (*bazeldnf.Config, error) {
	ignored := make(map[*api.Package]bool)
	ignoredNames := make(map[string]bool)
	for _, forceIgnoredPackage := range forceIgnored {
		ignored[forceIgnoredPackage] = true
		ignoredNames[forceIgnoredPackage.Name] = true
	}

	allPackages := make(map[*api.Package]*bazeldnf.RPM)
	repositories := make(map[string][]string)
	for _, installPackage := range install {
		repositories[installPackage.Repository.Name] = installPackage.Repository.Mirrors

		deps := make([]string, 0, len(installPackage.Format.Requires.Entries))
		for _, entry := range installPackage.Format.Requires.Entries {
			deps = append(deps, entry.Name)
		}

		slices.Sort(deps)

		integrity, err := installPackage.Checksum.Integrity()
		if err != nil {
			return nil, fmt.Errorf("Unable to read package %s integrity: %w", installPackage.Name, err)
		}

		allPackages[installPackage] = &bazeldnf.RPM{
			Id:           makeId(installPackage),
			Name:         installPackage.Name,
			Arch:         installPackage.Arch,
			Integrity:    integrity,
			URLs:         []string{installPackage.Location.Href},
			Repository:   installPackage.Repository.Name,
			Dependencies: deps,
		}
	}

	providers := collectProviders(forceIgnored, install)
	packageNames := sortedPackages(maps.Keys(allPackages))
	sortedPackages := make([]*bazeldnf.RPM, 0, len(packageNames))
	for _, name := range packageNames {
		pkg := allPackages[name]
		deps, err := collectDependencies(name, pkg.Dependencies, providers, ignored)
		if err != nil {
			return nil, err
		}

		pkg.Dependencies = make([]string, len(deps))
		for i, dep := range deps {
			pkg.Dependencies[i] = makeId(dep)
		}

		sortedPackages = append(sortedPackages, pkg)
	}

	lockFile := bazeldnf.Config{
		CommandLineArguments: cmdline,
		ForceIgnored:         sortedKeys(ignoredNames),
		RPMs:                 sortedPackages,
		Repositories:         repositories,
		Targets:              targets,
	}

	return &lockFile, nil
}

func collectProviders(pkgSets ...[]*api.Package) map[string][]*api.Package {
	providers := map[string][]*api.Package{}
	for _, pkgSet := range pkgSets {
		for _, pkg := range pkgSet {
			for _, entry := range pkg.Format.Provides.Entries {
				providers[entry.Name] = append(providers[entry.Name], pkg)
			}

			for _, entry := range pkg.Format.Files {
				providers[entry.Text] = append(providers[entry.Text], pkg)
			}
		}
	}

	return providers
}

func collectDependencies(pkg *api.Package, requires []string, providers map[string][]*api.Package, ignored map[*api.Package]bool) ([]*api.Package, error) {
	logrus.Debugf("Collecting dependencies for %s", pkg)
	depSet := make(map[*api.Package]bool)
	for _, req := range requires {
		logrus.Debugf("Resolving dependency %s", req)
		resolvedProviders, ok := providers[req]
		if !ok {
			return nil, fmt.Errorf("could not find provider for %s", req)
		}
		for _, provider := range resolvedProviders {
			logrus.Debugf("Found provider %s for %s", provider, req)
			if ignored[provider] {
				logrus.Debugf("Ignoring provider %s for %s", provider, req)
				continue
			}
			depSet[provider] = true
		}
	}

	deps := sortedPackages(maps.Keys(depSet))

	found := map[*api.Package]bool{pkg: true}

	// RPMs may have circular dependencies, even depend on themselves.
	// we need to ignore such dependencies
	nonCyclicDeps := make([]*api.Package, 0, len(deps))
	for _, dep := range deps {
		if found[dep] {
			continue
		}

		nonCyclicDeps = append(nonCyclicDeps, dep)
	}

	return nonCyclicDeps, nil
}
