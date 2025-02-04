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
	in         []string
	repofiles  []string
	out        string
	nobest     bool
	arch       string
	baseSystem string
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
				repos, err = repo.LoadRepoFiles(reduceopts.repofiles)
				if err != nil {
					return err
				}
			}
			_, involved, err := reducer.Resolve(repos, reduceopts.in, reduceopts.baseSystem, reduceopts.arch, required)
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

	reduceCmd.Flags().StringArrayVarP(&reduceopts.in, "input", "i", nil, "primary.xml of the repository")
	reduceCmd.Flags().StringVarP(&reduceopts.out, "output", "o", "debug.xml", "where to write the repository file")
	reduceCmd.Flags().StringVar(&reduceopts.baseSystem, "basesystem", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	reduceCmd.Flags().StringVarP(&reduceopts.arch, "arch", "a", "x86_64", "target architecture")
	reduceCmd.Flags().BoolVarP(&reduceopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	reduceCmd.Flags().StringArrayVarP(&reduceopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	// deprecated options
	reduceCmd.Flags().StringVarP(&reduceopts.baseSystem, "fedora-base-system", "f", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	reduceCmd.Flags().MarkDeprecated("fedora-base-system", "use --basesystem instead")
	reduceCmd.Flags().MarkShorthandDeprecated("fedora-base-system", "use --basesystem instead")
	reduceCmd.Flags().MarkShorthandDeprecated("nobest", "use --nobest instead")

	repo.AddCacheHelperFlags(reduceCmd)

	return reduceCmd
}
