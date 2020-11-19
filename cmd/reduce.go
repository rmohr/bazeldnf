package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type reduceOpts struct {
	in               []string
	repofile         string
	out              string
	lang             string
	nobest           bool
	arch             string
	fedoraBaseSystem string
}

var reduceopts = reduceOpts{}

func NewReduceCmd() *cobra.Command {

	reduceCmd := &cobra.Command{
		Use:   "reduce",
		Short: "debug command to produce trimmed down repos for testing or debugging purposes",
		Long: `reduces dependencies to all possible candidates for any dependency. This is mostly a debug command
which allow reducing huge rpm repos to a smaller problem set for debugging, removing all the unwanted noise of definitely unrelated packages.
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repos := &bazeldnf.Repositories{}
			if len(reduceopts.in) == 0 {
				var err error
				repos, err = repo.LoadRepoFile(reduceopts.repofile)
				if err != nil {
					return err
				}
			}
			repo := reducer.NewRepoReducer(repos, reduceopts.in, reduceopts.lang, reduceopts.fedoraBaseSystem, reduceopts.arch, ".bazeldnf")
			logrus.Info("Loading packages.")
			if err := repo.Load(); err != nil {
				return err
			}
			logrus.Info("Reduction of involved packages.")
			involved, err := repo.Resolve(required)
			if err != nil {
				return err
			}
			logrus.Info("Writing involved packages as a repo.")
			testrepo := &api.Repository{}
			for _, pkg := range involved {
				testrepo.Packages = append(testrepo.Packages, *pkg)
			}
			data, err := xml.MarshalIndent(testrepo, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal repository file: %v", err)
			}
			err = ioutil.WriteFile(reduceopts.out, data, 0666)
			if err != nil {
				return fmt.Errorf("failed to write repository file: %v", err)
			}
			return nil
		},
	}

	reduceCmd.PersistentFlags().StringArrayVarP(&reduceopts.in, "input", "i", nil, "primary.xml of the repository")
	reduceCmd.PersistentFlags().StringVarP(&reduceopts.out, "output", "o", "debug.xml", "where to write the repository file")
	reduceCmd.PersistentFlags().StringVarP(&reduceopts.fedoraBaseSystem, "fedora-base-system", "f", "fedora-release-container", "fedora base system to choose from (e.g. fedora-release-server, fedora-release-container, ...)")
	reduceCmd.PersistentFlags().StringVarP(&reduceopts.arch, "arch", "a", "x86_64", "target fedora architecture")
	reduceCmd.PersistentFlags().BoolVarP(&reduceopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	reduceCmd.PersistentFlags().StringVarP(&reduceopts.repofile, "repofile", "r", "repo.yaml", "repository information file. Will be used by default if no explicit inputs are provided.")
	return reduceCmd
}
