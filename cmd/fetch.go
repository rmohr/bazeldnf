package main

import (
	"io/ioutil"

	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
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

			repofile, err := ioutil.ReadFile(fetchopts.repofile)
			if err != nil {
				return err
			}
			repos := &bazeldnf.Repositories{}
			err = yaml.Unmarshal(repofile, repos)
			if err != nil {
				return err
			}
			return repo.NewRemoteRepoFetcher(repos.Repositories, ".bazeldnf").Fetch()
		},
	}

	fetchCmd.Flags().StringVarP(&fetchopts.repofile, "repofile", "r", "repo.yaml", "repository information file")
	return fetchCmd
}
