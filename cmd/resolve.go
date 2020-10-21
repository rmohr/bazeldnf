package main

import (
	"fmt"

	"github.com/rmohr/bazel-dnf/pkg/repoquery"
	"github.com/rmohr/bazel-dnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type resolveOpts struct {
	in            []string
	lang          string
	nobest        bool
	arch          string
	fedoraRelease string
}

var resolveopts = resolveOpts{}

func NewResolveCmd() *cobra.Command {

	resolveCmd := &cobra.Command{
		Use:   "resolve",
		Short: "resolves depencencies of the given packages",
		Long:  `resolves dependencies of the given packages with the assumption of a SCRATCH container as install target`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repo := repoquery.NewRepoQuerier(resolveopts.in, resolveopts.lang, resolveopts.fedoraRelease, resolveopts.arch)
			logrus.Info("Loading packages.")
			if err := repo.Load(); err != nil {
				return err
			}
			logrus.Info("Initial reduction of involved packages.")
			involved, err := repo.Resolve(required)
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
			err = solver.ConstructRequirements(required)
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

	resolveCmd.PersistentFlags().StringArrayVarP(&resolveopts.in, "input", "i", []string{"primary.xml"}, "primary.xml of the repository")
	resolveCmd.PersistentFlags().StringVarP(&resolveopts.fedoraRelease, "fedora-release", "f", "fedora-release-container", "fedora base system to choose from (e.g. fedora-release-server, fedora-release-container, ...)")
	resolveCmd.PersistentFlags().StringVarP(&resolveopts.lang, "lang", "l", "en", "language to use for locale decisions (like glibc-lang)")
	resolveCmd.PersistentFlags().StringVarP(&getopts.arch, "arch", "a", "x86_64", "target fedora architecture")
	resolveCmd.PersistentFlags().BoolVarP(&resolveopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	return resolveCmd
}
