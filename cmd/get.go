package main

import (
	"github.com/rmohr/bazel-dnf/pkg/repo"
	"github.com/spf13/cobra"
)

type GetOpts struct {
	arch    string
	version string
	out     string
}

var getopts = GetOpts{}

func NewGetCmd() *cobra.Command {

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "get the repository primary index for further processing",
		Long:  `Get primary repository files for different fedora versions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repo.NewRemoteRepoResolver(getopts.version, getopts.arch).Resolve(getopts.out)
		},
	}

	getCmd.PersistentFlags().StringVarP(&getopts.arch, "arch", "a", "x86_64", "target fedora architecture")
	getCmd.PersistentFlags().StringVarP(&getopts.version, "version", "v", "", "target fedora version")
	getCmd.PersistentFlags().StringVarP(&getopts.out, "output", "o", "primary.xml", "where to write the repository information")
	getCmd.MarkFlagRequired("arch")
	getCmd.MarkFlagRequired("version")
	return getCmd
}
