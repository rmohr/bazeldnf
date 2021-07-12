package main

import (
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
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
	fedoraBaseSystem string
	repofiles        []string
	workspace        string
	buildfile        string
	name             string
	public           bool
}

var rpmtreeopts = rpmtreeOpts{}

func NewRpmTreeCmd() *cobra.Command {

	rpmtreeCmd := &cobra.Command{
		Use:   "rpmtree",
		Short: "Writes a rpmtree rule and its rpmdependencies to bazel files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repos := &bazeldnf.Repositories{}
			for i, _ := range fetchopts.repofiles {
				tmp, err := repo.LoadRepoFile(rpmtreeopts.repofiles[i])
				if err != nil {
					return err
				}
				repos.Repositories = append(repos.Repositories, tmp.Repositories...)
			}
			repoReducer := reducer.NewRepoReducer(repos, nil, rpmtreeopts.lang, rpmtreeopts.fedoraBaseSystem, rpmtreeopts.arch, ".bazeldnf")
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
			err = solver.LoadInvolvedPackages(involved)
			if err != nil {
				return err
			}
			logrus.Info("Adding required packages to the rpmtreer.")
			err = solver.ConstructRequirements(matched)
			if err != nil {
				return err
			}
			logrus.Info("Solving.")
			install, _, err := solver.Resolve()
			if err != nil {
				return err
			}
			workspace, err := bazel.LoadWorkspace(rpmtreeopts.workspace)
			if err != nil {
				return err
			}
			build, err := bazel.LoadBuild(rpmtreeopts.buildfile)
			if err != nil {
				return err
			}
			bazel.AddRPMs(workspace, install, rpmtreeopts.arch)
			bazel.AddTree(rpmtreeopts.name, build, install, rpmtreeopts.arch, rpmtreeopts.public)
			bazel.PruneRPMs(build, workspace)
			logrus.Info("Writing bazel files.")
			err = bazel.WriteWorkspace(false, workspace, rpmtreeopts.workspace)
			if err != nil {
				return err
			}
			err = bazel.WriteBuild(false, build, rpmtreeopts.buildfile)
			if err != nil {
				return err
			}
			logrus.Info("Done.")

			return nil
		},
	}

	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.fedoraBaseSystem, "fedora-base-system", "f", "fedora-release-container", "fedora base system to choose from (e.g. fedora-release-server, fedora-release-container, ...)")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.arch, "arch", "a", "x86_64", "target fedora architecture")
	rpmtreeCmd.Flags().BoolVarP(&rpmtreeopts.nobest, "nobest", "n", false, "allow picking versions which are not the newest")
	rpmtreeCmd.Flags().BoolVarP(&rpmtreeopts.public, "public", "p", true, "if the rpmtree rule should be public")
	rpmtreeCmd.Flags().StringArrayVarP(&rpmtreeopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.name, "name", "", "", "rpmtree rule name")
	rpmtreeCmd.MarkFlagRequired("name")
	return rpmtreeCmd
}
