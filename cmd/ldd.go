package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/u-root/u-root/pkg/ldd"
)

type lddOpts struct {
	workspace string
	buildfile string
	name      string
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

			err = os.Setenv("LD_LIBRARY_PATH", tmpRoot)
			if err != nil {
				return err
			}
			dependencies, err := ldd.Ldd(objects)
			if err != nil {
				return err
			}
			fmt.Println(dependencies)

			logrus.Info("Done.")

			return nil
		},
	}

	lddCmd.Flags().StringVarP(&lddopts.tar, "input", "i", "", "Tar file with all dependencies")
	lddCmd.PersistentFlags().StringVarP(&lddopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	lddCmd.PersistentFlags().StringVarP(&lddopts.buildfile, "buildfile", "b", "rpm/BUILD.bazel", "Build file for RPMs")
	lddCmd.Flags().StringVarP(&lddopts.name, "name", "n", "", "rpmtree rule name")
	lddCmd.MarkFlagRequired("name")
	lddCmd.MarkFlagRequired("input")
	return lddCmd
}
