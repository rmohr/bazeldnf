package main

import (
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/spf13/cobra"
)

type InitOpts struct {
	arch string
	fc   string
	out  string
}

var initopts = InitOpts{}

func NewInitCmd() *cobra.Command {

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create basic repo.yaml files for fedora releases",
		Long:  `Create proper repo information with release- and update repos for fedora releases`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repo.NewRemoteInit(initopts.fc, initopts.arch, initopts.out).Init()
		},
	}

	initCmd.Flags().StringVarP(&initopts.arch, "arch", "a", "x86_64", "target architecture")
	initCmd.Flags().StringVar(&initopts.fc, "fc", "", "target fedora core release")
	initCmd.Flags().StringVarP(&initopts.out, "output", "o", "repo.yaml", "where to write the repository information")
	err := initCmd.MarkFlagRequired("fc")
	if err != nil {
		panic(err)
	}
	return initCmd
}
