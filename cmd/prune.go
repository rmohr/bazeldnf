package main

import (
	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type pruneOpts struct {
	workspace string
	toMacro   string
	buildfile string
}

var pruneopts = pruneOpts{}

func NewPruneCmd() *cobra.Command {

	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "prunes unused RPM dependencies",
		RunE: func(cmd *cobra.Command, required []string) error {
			build, err := bazel.LoadBuild(pruneopts.buildfile)
			if err != nil {
				return err
			}

			if pruneopts.toMacro == "" {
				workspace, err := bazel.LoadWorkspace(pruneopts.workspace)
				if err != nil {
					return err
				}
				bazel.PruneWorkspaceRPMs(build, workspace)
				err = bazel.WriteWorkspace(false, workspace, pruneopts.workspace)
				if err != nil {
					return err
				}
			} else {
				bzl, defname, err := bazel.ParseToMacro(pruneopts.toMacro)
				if err != nil {
					return err
				}
				bzlfile, err := bazel.LoadBzl(bzl)
				if err != nil {
					return err
				}
				bazel.PruneBzlfileRPMs(build, bzlfile, defname)
				err = bazel.WriteBzl(false, bzlfile, bzl)
				if err != nil {
					return err
				}
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
	pruneCmd.Flags().StringVarP(&pruneopts.toMacro, "to_macro", "", "", "Tells bazeldnf to write the RPMs to a macro in the given bzl file instead of the WORKSPACE file.The expected format is: macroFile%defName")
	pruneCmd.Flags().StringVarP(&pruneopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	return pruneCmd
}
