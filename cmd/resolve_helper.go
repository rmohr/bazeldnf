package main

import (
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type resolveHelperOpts struct {
	in               []string
	baseSystem       string
	arch             string
	nobest           bool
	ignoreMissing    bool
	forceIgnoreRegex []string
	onlyAllowRegex   []string
}

var resolvehelperopts = resolveHelperOpts{}

func resolve(repos *bazeldnf.Repositories, required []string) ([]*api.Package, []*api.Package, error) {
	matched, involved, err := reducer.Resolve(repos, resolvehelperopts.in, resolvehelperopts.baseSystem, resolvehelperopts.arch, required, resolvehelperopts.ignoreMissing)
	if err != nil {
		return nil, nil, err
	}

	if len(matched) == 0 {
		return nil, nil, nil
	}

	solver := sat.NewResolver()
	logrus.Info("Loading involved packages into the resolver.")
	model, err := solver.LoadInvolvedPackages(involved, matched, resolvehelperopts.forceIgnoreRegex, resolvehelperopts.onlyAllowRegex, resolvehelperopts.nobest)
	if err != nil {
		return nil, nil, err
	}
	logrus.Info("Solving.")
	install, _, forceIgnored, err := solver.Resolve(model)
	return install, forceIgnored, err
}

func addResolveHelperFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&resolvehelperopts.in, "input", "i", nil, "primary.xml of the repository")
	cmd.Flags().StringVar(&resolvehelperopts.baseSystem, "basesystem", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	cmd.Flags().StringVarP(&resolvehelperopts.arch, "arch", "a", "x86_64", "target architecture")
	cmd.Flags().BoolVarP(&resolvehelperopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	cmd.Flags().BoolVar(&resolvehelperopts.ignoreMissing, "ignore-missing", false, "ignore missing packages")
	cmd.Flags().StringArrayVar(&resolvehelperopts.forceIgnoreRegex, "force-ignore-with-dependencies", []string{}, "Packages matching these regex patterns will not be installed. Allows force-removing unwanted dependencies. Be careful, this can lead to hidden missing dependencies.")
	cmd.Flags().StringArrayVar(&resolvehelperopts.onlyAllowRegex, "only-allow", []string{}, "Packages matching these regex patterns may be installed. Allows scoping dependencies. Be careful, this can lead to hidden missing dependencies.")
	// deprecated options
	cmd.Flags().StringVarP(&resolvehelperopts.baseSystem, "fedora-base-system", "f", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	cmd.Flags().MarkDeprecated("fedora-base-system", "use --basesystem instead")
	cmd.Flags().MarkShorthandDeprecated("fedora-base-system", "use --basesystem instead")
	cmd.Flags().MarkShorthandDeprecated("nobest", "use --nobest instead")
}
