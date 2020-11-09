package main

import (
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/spf13/cobra"
)

type FetchOpts struct {
	repofile string
}

var fetchopts = &FetchOpts{}

func NewFetchCmd() *cobra.Command {

	fetchCmd := &cobra.Command{
		Use:   "fetch",
		Short: "Update repo metadata",
		Long:  `Update repo metadata`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := repo.LoadRepoFile(fetchopts.repofile)
			if err != nil {
				return err
			}
			return repo.NewRemoteRepoFetcher(repos.Repositories, ".bazeldnf").Fetch()
		},
	}

	fetchCmd.Flags().StringVarP(&fetchopts.repofile, "repofile", "r", "repo.yaml", "repository information file")
	return fetchCmd
}
