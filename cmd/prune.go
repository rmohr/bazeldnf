package main

import (
	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type pruneOpts struct {
	workspace string
	buildfile string
}

var pruneopts = pruneOpts{}

func NewPruneCmd() *cobra.Command {

	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "prunes unused RPM dependencies",
		RunE: func(cmd *cobra.Command, required []string) error {
			workspace, err := bazel.LoadWorkspace(pruneopts.workspace)
			if err != nil {
				return err
			}
			build, err := bazel.LoadBuild(pruneopts.buildfile)
			if err != nil {
				return err
			}
			bazel.PruneRPMs(build, workspace)
			err = bazel.WriteWorkspace(false, workspace, pruneopts.workspace)
			if err != nil {
				return err
			}
			err = bazel.WriteBuild(false, build, pruneopts.buildfile)
			if err != nil {
				return err
			}
			logrus.Info("Done.")
			return nil
		},
	}

	pruneCmd.Flags().StringVarP(&pruneopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	pruneCmd.Flags().StringVarP(&pruneopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	return pruneCmd
}
