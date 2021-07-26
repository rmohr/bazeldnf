package main

import (
	"os"

	"github.com/rmohr/bazeldnf/cmd/template"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/rmohr/bazeldnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type resolveOpts struct {
	in               []string
	lang             string
	nobest           bool
	arch             string
	baseSystem       string
	repofiles        []string
	forceIgnoreRegex []string
}

var resolveopts = resolveOpts{}

func NewResolveCmd() *cobra.Command {

	resolveCmd := &cobra.Command{
		Use:   "resolve",
		Short: "resolves depencencies of the given packages",
		Long:  `resolves dependencies of the given packages with the assumption of a SCRATCH container as install target`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repos := &bazeldnf.Repositories{}
			if len(resolveopts.in) == 0 {
				var err error
				repos, err = repo.LoadRepoFiles(resolveopts.repofiles)
				if err != nil {
					return err
				}
			}
			repo := reducer.NewRepoReducer(repos, resolveopts.in, resolveopts.lang, resolveopts.baseSystem, resolveopts.arch, ".bazeldnf")
			logrus.Info("Loading packages.")
			if err := repo.Load(); err != nil {
				return err
			}
			logrus.Info("Initial reduction of involved packages.")
			matched, involved, err := repo.Resolve(required)
			if err != nil {
				return err
			}
			solver := sat.NewResolver(resolveopts.nobest)
			logrus.Info("Loading involved packages into the resolver.")
			err = solver.LoadInvolvedPackages(involved, resolveopts.forceIgnoreRegex)
			if err != nil {
				return err
			}
			logrus.Info("Adding required packages to the resolver.")
			err = solver.ConstructRequirements(matched)
			if err != nil {
				return err
			}
			logrus.Info("Solving.")
			install, _, forceIgnored, err := solver.Resolve()
			if err != nil {
				return err
			}
			if err := template.Render(os.Stdout, install, forceIgnored); err != nil {
				return err
			}
			return nil
		},
	}

	resolveCmd.Flags().StringArrayVarP(&resolveopts.in, "input", "i", nil, "primary.xml of the repository")
	resolveCmd.Flags().StringVar(&resolveopts.baseSystem, "basesystem", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	resolveCmd.Flags().StringVarP(&resolveopts.arch, "arch", "a", "x86_64", "target architecture")
	resolveCmd.Flags().BoolVarP(&resolveopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	resolveCmd.Flags().StringArrayVarP(&resolveopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	resolveCmd.Flags().StringArrayVar(&resolveopts.forceIgnoreRegex, "force-ignore-with-dependencies", []string{}, "Packages matching these regex patterns will not be installed. Allows force-removing unwanted dependencies. Be careful, this can lead to hidden missing dependencies.")
	// deprecated options
	resolveCmd.Flags().StringVarP(&resolveopts.baseSystem, "fedora-base-system", "f", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	resolveCmd.Flags().MarkDeprecated("fedora-base-system", "use --basesystem instead")
	resolveCmd.Flags().MarkShorthandDeprecated("fedora-base-system", "use --basesystem instead")
	return resolveCmd
}
