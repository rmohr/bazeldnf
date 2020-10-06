package main

import (
	"github.com/rmohr/bazel-dnf/pkg/repo"
	"github.com/spf13/cobra"
)

var arch string
var version string
var out string

func NewGetCmd() *cobra.Command {

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "get the repository primary index for further processing",
		Long: `Get primary repository files for different fedora versions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repo.NewRemoteRepoResolver(version, arch).Resolve(out)
		},
	}

	getCmd.PersistentFlags().StringVarP(&arch, "arch", "a", "", "target fedora architecture")
	getCmd.PersistentFlags().StringVarP(&version, "version", "v", "", "target fedora version")
	getCmd.PersistentFlags().StringVarP(&out, "output", "o", "primary.xml", "where to write the repository information")
	getCmd.MarkFlagRequired("arch")
	getCmd.MarkFlagRequired("version")
	return getCmd
}
