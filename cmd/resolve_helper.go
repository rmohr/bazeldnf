package main

import (
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/sat"
	"github.com/sirupsen/logrus"
)

type resolveHelperOpts struct {
	in               []string
	baseSystem       string
	arch             string
	nobest           bool
	forceIgnoreRegex []string
	onlyAllowRegex   []string
}

var resolvehelperopts = resolveHelperOpts{}

func resolve(repos *bazeldnf.Repositories, required []string) ([]*api.Package, []*api.Package, error) {
	matched, involved, err := reducer.Resolve(repos, resolvehelperopts.in, resolvehelperopts.baseSystem, resolvehelperopts.arch, required)
	if err != nil {
		return nil, nil, err
	}
	solver := sat.NewResolver(resolvehelperopts.nobest)
	logrus.Info("Loading involved packages into the resolver.")
	err = solver.LoadInvolvedPackages(involved, resolvehelperopts.forceIgnoreRegex, resolvehelperopts.onlyAllowRegex)
	if err != nil {
		return nil, nil, err
	}
	logrus.Info("Adding required packages to the resolver.")
	err = solver.ConstructRequirements(matched)
	if err != nil {
		return nil, nil, err
	}
	logrus.Info("Solving.")
	install, _, forceIgnored, err := solver.Resolve()
	return install, forceIgnored, err
}

