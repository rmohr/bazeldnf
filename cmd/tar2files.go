package main

import (
	"archive/tar"
	"fmt"
	"os"

	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/spf13/cobra"
)

var filePrefix string
var tarFile string

func NewTar2FilesCmd() *cobra.Command {
	tarCmd := &cobra.Command{
		Use:   "tar2files",
		Short: "extract certain files in a given directory from a tar archive",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tarStream := os.Stdin
			if tarFile != "" {
				tarStream, err = os.Open(tarFile)
				if err != nil {
					return fmt.Errorf("could not open rpm at %s: %v", tarFile, err)
				}
				err = rpm.PrefixFilter(filePrefix, tar.NewReader(tarStream), args)
				if err != nil {
					return fmt.Errorf("could not convert rpm at %s: %v", tarFile, err)
				}
			} else {
				err = rpm.PrefixFilter(filePrefix, tar.NewReader(tarStream), args)
				if err != nil {
					return fmt.Errorf("could not convert rpm : %v", err)
				}
			}
			return nil
		},
	}

	tarCmd.PersistentFlags().StringVarP(&tarFile, "input", "i", "", "location from where to read the tar file (defaults to stdin)")
	tarCmd.PersistentFlags().StringVar(&filePrefix, "file-prefix", "", "only keep files with this directory prefix")
	return tarCmd
}
