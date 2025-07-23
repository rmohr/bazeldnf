package main

import (
	"os"

	"github.com/rmohr/bazeldnf/cmd/template"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/spf13/cobra"
)

type resolveOpts struct {
	repofiles []string
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
			if len(resolvehelperopts.in) == 0 {
				var err error
				repos, err = repo.LoadRepoFiles(resolveopts.repofiles)
				if err != nil {
					return err
				}
			}

			if resolvehelperopts.baseSystem == "scratch" {
				resolvehelperopts.baseSystem = ""
			}

			install, forceIgnored, err := resolve(repos, required)
			if err != nil {
				return err
			}

			if err := template.Render(os.Stdout, install, forceIgnored); err != nil {
				return err
			}
			return nil
		},
	}

	resolveCmd.Flags().StringArrayVarP(&resolveopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")

	repo.AddCacheHelperFlags(resolveCmd)
	addResolveHelperFlags(resolveCmd)

	return resolveCmd
}
