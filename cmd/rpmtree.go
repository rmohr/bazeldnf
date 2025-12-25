package main

import (
	"os"

	"github.com/bazelbuild/buildtools/build"
	"github.com/rmohr/bazeldnf/cmd/template"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type rpmtreeOpts struct {
	repofiles   []string
	workspace   string
	toMacro     string
	buildfile   string
	configname  string
	lockfile    string
	name        string
	public      bool
	compression string
}

var rpmtreeopts = rpmtreeOpts{}

type Handler interface {
	Process(pkgs []*api.Package, buildfile *build.File) error
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

func (h *MacroHandler) Process(pkgs []*api.Package, buildfile *build.File) error {
	if err := bazel.AddBzlfileRPMs(h.bzlfile, h.defName, pkgs); err != nil {
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

func (h *WorkspaceHandler) Process(pkgs []*api.Package, buildfile *build.File) error {
	if err := bazel.AddWorkspaceRPMs(h.workspacefile, pkgs); err != nil {
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
			RPMs: []*bazeldnf.RPM{},
		},
	}, nil
}

func (h *LockFileHandler) Process(pkgs []*api.Package, buildfile *build.File) error {
	return bazel.AddConfigRPMs(h.config, pkgs)
}

func (h *LockFileHandler) Write() error {
	return bazel.WriteLockFile(h.config, h.filename)
}

func newHandler() (Handler, string, error) {
	if rpmtreeopts.toMacro != "" {
		handler, err := NewMacroHandler(rpmtreeopts.toMacro)
		return handler, "", err
	} else if rpmtreeopts.lockfile != "" {
		handler, err := NewLockFileHandler(
			rpmtreeopts.configname,
			rpmtreeopts.lockfile,
		)
		return handler, rpmtreeopts.configname, err
	}

	handler, err := NewWorkspaceHandler(rpmtreeopts.workspace)
	return handler, "", err
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
			install, forceIgnored, err := resolve(repos, required)
			if err != nil {
				return err
			}

			handler, configname, err := newHandler()
			if err != nil {
				return err
			}

			build, err := bazel.LoadBuild(rpmtreeopts.buildfile)
			if err != nil {
				return err
			}
			bazel.AddTree(rpmtreeopts.name, configname, build, install, rpmtreeopts.public, rpmtreeopts.compression)

			if err := handler.Process(install, build); err != nil {
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

	rpmtreeCmd.Flags().BoolVarP(&rpmtreeopts.public, "public", "p", true, "if the rpmtree rule should be public")
	rpmtreeCmd.Flags().StringArrayVarP(&rpmtreeopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.toMacro, "to-macro", "", "", "Tells bazeldnf to write the RPMs to a macro in the given bzl file instead of the WORKSPACE file. The expected format is: macroFile%defName")
	rpmtreeCmd.Flags().StringVarP(&rpmtreeopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.configname, "configname", "rpms", "config name to use in lockfile")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.lockfile, "lockfile", "", "lockfile for RPMs")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.name, "name", "", "rpmtree rule name")
	rpmtreeCmd.Flags().StringVar(&rpmtreeopts.compression, "compression", "", "Compression algorithm to use on resulting archive (e.g., gzip)")
	rpmtreeCmd.MarkFlagRequired("name")

	repo.AddCacheHelperFlags(rpmtreeCmd)
	addResolveHelperFlags(rpmtreeCmd)

	return rpmtreeCmd
}
