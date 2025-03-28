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

func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}

func toConfig(install, forceIgnored []*api.Package, targets []string, cmdline []string) (*bazeldnf.Config, error) {
	ignored := make(map[string]bool)
	for _, forceIgnoredPackage := range forceIgnored {
		ignored[forceIgnoredPackage.Name] = true
	}

	allPackages := make(map[string]*bazeldnf.RPM)
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

		allPackages[installPackage.Name] = &bazeldnf.RPM{
			Name:         installPackage.Name,
			Integrity:    integrity,
			URLs:         []string{installPackage.Location.Href},
			Repository:   installPackage.Repository.Name,
			Dependencies: deps,
		}
	}

	providers := collectProviders(forceIgnored, install)
	slices.Sort(targets)
	sortedPackages := make([]*bazeldnf.RPM, 0, len(allPackages))

	requested := make(map[string]bool)
	for _, pkg := range targets {
		requested[pkg] = true
	}

	// make sure all requested packages have their full dependency tree set
	for _, name := range sortedKeys(allPackages) {
		pkg := allPackages[name]
		if _, requested := requested[name]; !requested {
			continue
		}
		deps, err := collectDependencies(name, pkg.Dependencies, providers, ignored, allPackages)
		if err != nil {
			return nil, err
		}

		pkg.SetDependencies(deps)

		sortedPackages = append(sortedPackages, pkg)
	}

	// now for non requested packages make sure we don't get cycles
	for _, name := range sortedKeys(allPackages) {
		pkg := allPackages[name]
		if _, requested := requested[name]; requested {
			continue
		}

		pkg.SetDependencies(nil)

		sortedPackages = append(sortedPackages, pkg)
	}

	slices.SortFunc(sortedPackages, func(a, b *bazeldnf.RPM) int {
		return cmp.Compare(a.Name, b.Name)
	})

	lockFile := bazeldnf.Config{
		CommandLineArguments: cmdline,
		ForceIgnored:         sortedKeys(ignored),
		RPMs:                 sortedPackages,
		Repositories:         repositories,
		Targets:              targets,
	}

	return &lockFile, nil
}

func collectProviders(pkgSets ...[]*api.Package) map[string]string {
	providers := map[string]string{}
	for _, pkgSet := range pkgSets {
		for _, pkg := range pkgSet {
			for _, entry := range pkg.Format.Provides.Entries {
				providers[entry.Name] = pkg.Name
			}

			for _, entry := range pkg.Format.Files {
				providers[entry.Text] = pkg.Name
			}
		}
	}

	return providers
}

func collectDependencies(pkg string, requires []string, providers map[string]string, ignored map[string]bool, allPackages map[string]*bazeldnf.RPM) ([]string, error) {
	logrus.Debugf("Collecting dependencies for %s", pkg)
	depSet := make(map[string]bool)
	explored := make(map[string]bool)
	for len(requires) > 0 {
		req := requires[0]
		requires = requires[1:]
		logrus.Debugf("Processing dependency %s, pending %d", req, len(requires))
		if explored[req] {
			logrus.Debugf("Ignoring already explored %s", req)
			continue
		}
		explored[req] = true
		if ignored[req] {
			logrus.Debugf("Ignoring dependency %s", req)
			continue
		}
		logrus.Debugf("Resolving dependency %s", req)
		provider, ok := providers[req]
		if !ok {
			return nil, fmt.Errorf("could not find provider for %s", req)
		}
		if ignored[provider] {
			logrus.Debugf("Ignoring provider %s", provider)
			continue
		}
		depSet[provider] = true
		requires = append(requires, allPackages[provider].Dependencies...)
	}
	return sortedKeys(depSet), nil
}

func removeCyclicDependencies(targets []string, allPackages []*bazeldnf.RPM) []*bazeldnf.RPM {
	allPackagesMap := make(map[string]*bazeldnf.RPM)

	for _, installPackage := range allPackages {
		allPackagesMap[installPackage.Name] = installPackage
	}

	for _, target := range targets {
		visitedMap := make(map[string]bool)
		recursionStack := make(map[string]bool)

		removeCyclicDependenciesHelper(allPackagesMap, target, visitedMap, recursionStack)
	}

	return allPackages
}

func removeCyclicDependenciesHelper(allPackages map[string]*bazeldnf.RPM, pkg string, visitedMap, recursionStack map[string]bool) bool {
	/*
	 * This is a recursive function that removes cyclic dependencies from the
	 * dependency graph in the case cycles are found
	 */
	visitedMap[pkg] = true
	recursionStack[pkg] = true

	if _, ok := allPackages[pkg]; !ok {
		return false
	}

	if allPackages[pkg].Dependencies == nil {
		return false
	}

	cleanDependencies := make([]string, 0, len(allPackages[pkg].Dependencies))

	for _, dep := range allPackages[pkg].Dependencies {
		if _, visited := visitedMap[dep]; !visited {
			if removeCyclicDependenciesHelper(allPackages, dep, visitedMap, recursionStack) {
				// ignore cycle
				logrus.Debugf("Ignoring cyclic dependency %s -> %s", pkg, dep)
				continue
			}
		} else if _, recursed := recursionStack[dep]; recursed {
			// ignore cycle
			logrus.Debugf("Ignoring cyclic dependency in recursion stack %s -> %s", pkg, dep)
			continue
		}
		cleanDependencies = append(cleanDependencies, dep)
	}

	recursionStack[pkg] = false

	newPkg := allPackages[pkg].Clone()
	newPkg.SetDependencies(cleanDependencies)
	allPackages[pkg] = newPkg

	return false
}
