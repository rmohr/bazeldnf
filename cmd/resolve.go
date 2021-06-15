package main

import (
	"fmt"

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
	fedoraBaseSystem string
	repofile         string
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
				repos, err = repo.LoadRepoFile(resolveopts.repofile)
				if err != nil {
					return err
				}
			}
			repo := reducer.NewRepoReducer(repos, resolveopts.in, resolveopts.lang, resolveopts.fedoraBaseSystem, resolveopts.arch, ".bazeldnf")
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
			err = solver.LoadInvolvedPackages(involved)
			if err != nil {
				return err
			}
			logrus.Info("Adding required packages to the resolver.")
			err = solver.ConstructRequirements(matched)
			if err != nil {
				return err
			}
			logrus.Info("Solving.")
			install, _, err := solver.Resolve()
			if err != nil {
				return err
			}
			fmt.Println(install)
			fmt.Println(len(install))
			logrus.Info("Done.")
			return nil
		},
	}

	resolveCmd.PersistentFlags().StringArrayVarP(&resolveopts.in, "input", "i", nil, "primary.xml of the repository")
	resolveCmd.PersistentFlags().StringVarP(&resolveopts.fedoraBaseSystem, "fedora-base-system", "f", "fedora-release-container", "fedora base system to choose from (e.g. fedora-release-server, fedora-release-container, ...)")
	resolveCmd.PersistentFlags().StringVarP(&resolveopts.arch, "arch", "a", "x86_64", "target fedora architecture")
	resolveCmd.PersistentFlags().BoolVarP(&resolveopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	resolveCmd.PersistentFlags().StringVarP(&resolveopts.repofile, "repofile", "r", "repo.yaml", "repository information file. Will be used by default if no explicit inputs are provided.")
	return resolveCmd
}
