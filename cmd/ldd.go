package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/ldd"
	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type lddOpts struct {
	workspace string
	buildfile string
	name      string
	rpmtree   string
	tar       string
	public    bool
}

var lddopts = lddOpts{}

func NewLddCmd() *cobra.Command {

	lddCmd := &cobra.Command{
		Use:   "ldd",
		Short: "Determine shared library dependencies of binaries",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, objects []string) error {
			tmpRoot, err := ioutil.TempDir("", "bazeldnf-ldd")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpRoot)

			err = rpm.Untar(tmpRoot, lddopts.tar)
			if err != nil {
				return err
			}

			for i, _ := range objects {
				objects[i] = filepath.Join(tmpRoot, objects[i])
			}

			files := []string{}
			dependencies, err := ldd.Resolve(objects, []string{filepath.Join(tmpRoot, "/usr/lib64"), filepath.Join(tmpRoot, "/usr/lib")})
			if err != nil {
				return err
			}
			for _, dep := range dependencies {
				if strings.HasPrefix(dep, tmpRoot) {
					files = append(files, strings.TrimPrefix(dep, tmpRoot))
				}
			}
			err = filepath.Walk(filepath.Join(tmpRoot, "/usr/include"),
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						files = append(files, strings.TrimPrefix(path, tmpRoot))
					}
					return nil
				})
			if err != nil {
				log.Println(err)
				return err
			}
			build, err := bazel.LoadBuild(lddopts.buildfile)
			if err != nil {
				return err
			}
			bazel.AddTar2Files(lddopts.name, lddopts.rpmtree, build, filterFiles(files), lddopts.public)
			err = bazel.WriteBuild(false, build, lddopts.buildfile)
			if err != nil {
				return err
			}

			logrus.Info("Done.")

			return nil
		},
	}

	lddCmd.Flags().StringVarP(&lddopts.tar, "input", "i", "", "Tar file with all dependencies")
	lddCmd.Flags().StringVarP(&lddopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	lddCmd.Flags().StringVarP(&lddopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	lddCmd.Flags().BoolVarP(&lddopts.public, "public", "p", true, "if the tar2files rule should be public")
	lddCmd.Flags().StringVarP(&lddopts.name, "name", "n", "", "tar2files rule name")
	lddCmd.Flags().StringVarP(&lddopts.rpmtree, "rpmtree", "r", "", "rpmtree rule name")
	lddCmd.MarkFlagRequired("name")
	lddCmd.MarkFlagRequired("input")
	// deprecated options
	lddCmd.Flags().MarkShorthandDeprecated("name", "use --name instead")
	lddCmd.Flags().MarkShorthandDeprecated("rpmtree", "use --rpmtree instead")
	return lddCmd
}
