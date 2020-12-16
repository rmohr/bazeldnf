package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/u-root/u-root/pkg/ldd"
)

type lddOpts struct {
	workspace string
	buildfile string
	name      string
	rpmtree   string
	tar       string
}

var lddopts = lddOpts{}

func NewlddCmd() *cobra.Command {

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

			err = os.Setenv("LD_LIBRARY_PATH", filepath.Join(tmpRoot, "/usr/lib64"))
			if err != nil {
				return err
			}

			files := []string{}
			dependencies, err := ldd.Ldd(objects)
			if err != nil {
				return err
			}
			for _, dep := range dependencies {
				if strings.HasPrefix(dep.FullName, tmpRoot) {
					files = append(files, strings.TrimPrefix(dep.FullName, tmpRoot))
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
				return err
				log.Println(err)
			}
			if err != nil {
				return err
			}
			build, err := bazel.LoadBuild(lddopts.buildfile)
			if err != nil {
				return err
			}
			bazel.AddTar2Files(lddopts.name, lddopts.rpmtree, build, filterFiles(files), rpmtreeopts.public)
			err = bazel.WriteBuild(false, build, rpmtreeopts.buildfile)
			if err != nil {
				return err
			}

			logrus.Info("Done.")

			return nil
		},
	}

	lddCmd.Flags().StringVarP(&lddopts.tar, "input", "i", "", "Tar file with all dependencies")
	lddCmd.PersistentFlags().StringVarP(&lddopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	lddCmd.PersistentFlags().StringVarP(&lddopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	lddCmd.Flags().StringVarP(&lddopts.name, "name", "n", "", "tar2files rule name")
	lddCmd.Flags().StringVarP(&lddopts.rpmtree, "rpmtree", "r", "", "rpmtree rule name")
	lddCmd.MarkFlagRequired("name")
	lddCmd.MarkFlagRequired("input")
	return lddCmd
}
