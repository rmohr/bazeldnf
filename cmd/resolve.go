package main

import (
	"fmt"

	"github.com/rmohr/bazel-dnf/pkg/repoquery"
	"github.com/rmohr/bazel-dnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var in []string
var lang string


func NewResolveCmd() *cobra.Command {

	resolveCmd := &cobra.Command{
		Use:   "resolve",
		Short: "resolves depencencies of the given packages",
		Long: `resolves dependencies of the given packages with the assumption of a SCRATCH container as install target`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repo := repoquery.NewRepoQuerier(in, lang)
			logrus.Info("Loading packages.")
			if err := repo.Load(); err != nil {
				return err
			}
			logrus.Info("Initial reduction of involved packages.")
			involved, err := repo.Resolve(required)
			if err != nil {
				return err
			}
			solver := sat.NewResolver()
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
			install, excluded, err := solver.Resolve()
			fmt.Println(install)
			fmt.Println(excluded)
			logrus.Info("Done.")
			fmt.Println(err)
			mus, err := solver.MUS()
			fmt.Println(err)
			fmt.Println(mus.CNF())

			return nil
		},
	}

	resolveCmd.PersistentFlags().StringArrayVarP(&in, "input", "i", []string{"primary.xml"}, "primary.xml of the repository")
	resolveCmd.PersistentFlags().StringVarP(&lang, "lang", "l", "en", "language to use for locale decisions (like glibc-lang)")
	return resolveCmd
}
