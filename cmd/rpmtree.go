package main

import (
	"os"

	"github.com/bazelbuild/buildtools/build"
	"github.com/rmohr/bazeldnf/cmd/template"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/rmohr/bazeldnf/pkg/sat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type rpmtreeOpts struct {
	nobest           bool
	arch             string
	baseSystem       string
	repofiles        []string
	workspace        string
	toMacro          string
	buildfile        string
	configname       string
	lockfile         string
	name             string
	public           bool
	forceIgnoreRegex []string
	onlyAllowRegex   []string
}

var rpmtreeopts = rpmtreeOpts{}

type Handler interface {
	Process(pkgs []*api.Package, arch string, buildfile *build.File) error
	Write() error
}

type MacroHandler struct {
	bzl, defName string
	bzlfile      *build.File
}

func NewMacroHandler(toMacro string) (Handler, error) {
	bzl, defName, err := bazel.ParseMacro(rpmtreeopts.toMacro)

	if err != nil {
		return nil, err
	}

	bzlfile, err := bazel.LoadBzl(bzl)
	if err != nil {
		return nil, err
	}

	return &MacroHandler{
		bzl:     bzl,
		bzlfile: bzlfile,
		defName: defName,
	}, nil
}

func (h *MacroHandler) Process(pkgs []*api.Package, arch string, buildfile *build.File) error {
	if err := bazel.AddBzlfileRPMs(h.bzlfile, h.defName, pkgs, arch); err != nil {
		return err
	}

	bazel.PruneBzlfileRPMs(buildfile, h.bzlfile, h.defName)
	return nil
}

func (h *MacroHandler) Write() error {
	return bazel.WriteBzl(false, h.bzlfile, h.bzl)
}

type WorkspaceHandler struct {
	workspace     string
	workspacefile *build.File
}

func NewWorkspaceHandler(workspace string) (Handler, error) {
	workspacefile, err := bazel.LoadWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return &WorkspaceHandler{
		workspace:     workspace,
		workspacefile: workspacefile,
	}, nil
}

func (h *WorkspaceHandler) Process(pkgs []*api.Package, arch string, buildfile *build.File) error {
	if err := bazel.AddWorkspaceRPMs(h.workspacefile, pkgs, arch); err != nil {
		return err
	}

	bazel.PruneWorkspaceRPMs(buildfile, h.workspacefile)
	return nil
}

func (h *WorkspaceHandler) Write() error {
	return bazel.WriteWorkspace(false, h.workspacefile, h.workspace)
}

type LockFileHandler struct {
	filename string
	config   *bazeldnf.Config
}

func NewLockFileHandler(configname, filename string) (Handler, error) {
	return &LockFileHandler{
		filename: filename,
		config: &bazeldnf.Config{
			Name: configname,
			RPMs: []bazeldnf.RPM{},
		},
	}, nil
}

func (h *LockFileHandler) Process(pkgs []*api.Package, arch string, buildfile *build.File) error {
	return bazel.AddConfigRPMs(h.config, pkgs, arch)
}

func (h *LockFileHandler) Write() error {
	return bazel.WriteLockFile(h.config, h.filename)
}

func NewRpmTreeCmd() *cobra.Command {

	rpmtreeCmd := &cobra.Command{
		Use:   "rpmtree",
		Short: "Writes a rpmtree rule and its rpmdependencies to bazel files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repos, err := repo.LoadRepoFiles(rpmtreeopts.repofiles)
			if err != nil {
				return err
			}
			repoReducer := reducer.NewRepoReducer(repos, nil, rpmtreeopts.baseSystem, rpmtreeopts.arch, repo.NewCacheHelper())
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
			err = solver.LoadInvolvedPackages(involved, rpmtreeopts.forceIgnoreRegex, rpmtreeopts.onlyAllowRegex)
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

			var handler Handler
			var configname string

			if rpmtreeopts.toMacro != "" {
				handler, err = NewMacroHandler(rpmtreeopts.toMacro)
			} else if rpmtreeopts.lockfile != "" {
				configname = rpmtreeopts.configname
				handler, err = NewLockFileHandler(
					rpmtreeopts.configname,
					rpmtreeopts.lockfile,
				)
			} else {
				handler, err = NewWorkspaceHandler(rpmtreeopts.workspace)
			}

			if err != nil {
				return err
			}

			build, err := bazel.LoadBuild(rpmtreeopts.buildfile)
			if err != nil {
				return err
			}
			bazel.AddTree(rpmtreeopts.name, configname, build, install, rpmtreeopts.arch, rpmtreeopts.public)

			if err := handler.Process(install, rpmtreeopts.arch, build); err != nil {
				return err
			}

			logrus.Info("Writing bazel files.")
			err = handler.Write()
			if err != nil {
				return err
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
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.toMacro, "to-macro", "", "", "Tells bazeldnf to write the RPMs to a macro in the given bzl file instead of the WORKSPACE file. The expected format is: macroFile%defName")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.configname, "configname", "rpms", "config name to use in lockfile")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.lockfile, "lockfile", "", "lockfile for RPMs")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.name, "name", "", "rpmtree rule name")
	rpmtreeCmd.Flags().StringArrayVar(&rpmtreeopts.forceIgnoreRegex, "force-ignore-with-dependencies", []string{}, "Packages matching these regex patterns will not be installed. Allows force-removing unwanted dependencies. Be careful, this can lead to hidden missing dependencies.")
	rpmtreeCmd.Flags().StringArrayVar(&rpmtreeopts.onlyAllowRegex, "only-allow", []string{}, "Packages matching these regex patterns may be installed. Allows scoping dependencies. Be careful, this can lead to hidden missing dependencies.")
	rpmtreeCmd.MarkFlagRequired("name")
	// deprecated options
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.baseSystem, "fedora-base-system", "f", "fedora-release-container", "base system to use (e.g. fedora-release-server, centos-stream-release, ...)")
	rpmtreeCmd.Flags().MarkDeprecated("fedora-base-system", "use --basesystem instead")
	rpmtreeCmd.Flags().MarkShorthandDeprecated("fedora-base-system", "use --basesystem instead")
	rpmtreeCmd.Flags().MarkShorthandDeprecated("nobest", "use --nobest instead")
	repo.AddCacheHelperFlags(rpmtreeCmd)

	return rpmtreeCmd
}
