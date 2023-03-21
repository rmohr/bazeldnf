package main

import (
	"os"

	"github.com/bazelbuild/buildtools/build"
	"github.com/rmohr/bazeldnf/cmd/template"
	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/rmohr/bazeldnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type rpmtreeOpts struct {
	lang             string
	nobest           bool
	arch             string
	baseSystem       string
	repofiles        []string
	workspace        string
	toMacro          string
	buildfile        string
	name             string
	public           bool
	forceIgnoreRegex []string
}

var rpmtreeopts = rpmtreeOpts{}

func NewRpmTreeCmd() *cobra.Command {

	rpmtreeCmd := &cobra.Command{
		Use:   "rpmtree",
		Short: "Writes a rpmtree rule and its rpmdependencies to bazel files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			writeToMacro := rpmtreeopts.toMacro != ""

			repos, err := repo.LoadRepoFiles(rpmtreeopts.repofiles)
			if err != nil {
				return err
			}
			repoReducer := reducer.NewRepoReducer(repos, nil, rpmtreeopts.lang, rpmtreeopts.baseSystem, rpmtreeopts.arch, ".bazeldnf")
			logrus.Info("Loading packages.")
			if err := repoReducer.Load(); err != nil {
				return err
			}
			logrus.Info("Initial reduction of involved packages.")
			matched, involved, err := repoReducer.Resolve(required)
			if err != nil {
				return err
			}
			solver := sat.NewResolver(rpmtreeopts.nobest)
			logrus.Info("Loading involved packages into the rpmtreer.")
			err = solver.LoadInvolvedPackages(involved, rpmtreeopts.forceIgnoreRegex)
			if err != nil {
				return err
			}
			logrus.Info("Adding required packages to the rpmtreer.")
			err = solver.ConstructRequirements(matched)
			if err != nil {
				return err
			}
			logrus.Info("Solving.")
			install, _, forceIgnored, err := solver.Resolve()
			if err != nil {
				return err
			}
			workspace, err := bazel.LoadWorkspace(rpmtreeopts.workspace)
			if err != nil {
				return err
			}
			var bzlfile *build.File
			var bzl, defName string
			if writeToMacro {
				bzl, defName, err = bazel.ParseToMacro(rpmtreeopts.toMacro)
				if err != nil {
					return err
				}
				bzlfile, err = bazel.LoadBzl(bzl)
				if err != nil {
					return err
				}
			}
			build, err := bazel.LoadBuild(rpmtreeopts.buildfile)
			if err != nil {
				return err
			}
			if writeToMacro {
				err = bazel.AddBzlfileRPMs(bzlfile, defName, install, rpmtreeopts.arch)
				if err != nil {
					return err
				}
			} else {
				err = bazel.AddWorkspaceRPMs(workspace, install, rpmtreeopts.arch)
				if err != nil {
					return err
				}
			}
			bazel.AddTree(rpmtreeopts.name, build, install, rpmtreeopts.arch, rpmtreeopts.public)
			if writeToMacro {
				bazel.PruneBzlfileRPMs(build, bzlfile, defName)
			} else {
				bazel.PruneWorkspaceRPMs(build, workspace)
			}
			logrus.Info("Writing bazel files.")
			err = bazel.WriteWorkspace(false, workspace, rpmtreeopts.workspace)
			if err != nil {
				return err
			}
			if writeToMacro {
				err = bazel.WriteBzl(false, bzlfile, bzl)
				if err != nil {
					return err
				}
			}
			err = bazel.WriteBuild(false, build, rpmtreeopts.buildfile)
			if err != nil {
				return err
			}
			if err := template.Render(os.Stdout, install, forceIgnored); err != nil {
				return err
			}

			return nil
		},
	}

	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.baseSystem, "basesystem", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.arch, "arch", "a", "x86_64", "target architecture")
	rpmtreeCmd.Flags().BoolVarP(&rpmtreeopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	rpmtreeCmd.Flags().BoolVarP(&rpmtreeopts.public, "public", "p", true, "if the rpmtree rule should be public")
	rpmtreeCmd.Flags().StringArrayVarP(&rpmtreeopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.toMacro, "to_macro", "", "", "Tells bazeldnf to write the RPMs to a macro in the given bzl file instead of the WORKSPACE file.The expected format is: macroFile%defName")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.name, "name", "", "rpmtree rule name")
	rpmtreeCmd.Flags().StringArrayVar(&rpmtreeopts.forceIgnoreRegex, "force-ignore-with-dependencies", []string{}, "Packages matching these regex patterns will not be installed. Allows force-removing unwanted dependencies. Be careful, this can lead to hidden missing dependencies.")
	rpmtreeCmd.MarkFlagRequired("name")
	// deprecated options
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.baseSystem, "fedora-base-system", "f", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	rpmtreeCmd.Flags().MarkDeprecated("fedora-base-system", "use --basesystem instead")
	rpmtreeCmd.Flags().MarkShorthandDeprecated("fedora-base-system", "use --basesystem instead")
	rpmtreeCmd.Flags().MarkShorthandDeprecated("nobest", "use --nobest instead")
	return rpmtreeCmd
}
