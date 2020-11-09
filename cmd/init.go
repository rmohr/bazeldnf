package main

import (
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/spf13/cobra"
)

type GetOpts struct {
	arch string
	fc   string
	out  string
}

var getopts = GetOpts{}

func NewGetCmd() *cobra.Command {

	getCmd := &cobra.Command{
		Use:   "init",
		Short: "Create basic repo.yaml files for fedora releases",
		Long:  `Create proper repo information with release- and update repos for fedora releases`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repo.NewRemoteInit(getopts.fc, getopts.arch, getopts.out).Init()
		},
	}

	getCmd.Flags().StringVarP(&getopts.arch, "arch", "a", "x86_64", "target fedora architecture")
	getCmd.Flags().StringVarP(&getopts.fc, "fc", "", "", "target fedora core release")
	getCmd.Flags().StringVarP(&getopts.out, "output", "o", "repo.yaml", "where to write the repository information")
	err := getCmd.MarkFlagRequired("fc")
	if err != nil {
		panic(err)
	}
	return getCmd
}
