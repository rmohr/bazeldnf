package main

import (
	"archive/tar"
	"fmt"
	"os"

	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/spf13/cobra"
)

type tar2filesOpts struct {
	filePrefix string
	tarFile    string
}

var tar2filesopts = tar2filesOpts{}

func NewTar2FilesCmd() *cobra.Command {
	tar2filesCmd := &cobra.Command{
		Use:   "tar2files",
		Short: "extract certain files in a given directory from a tar archive",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tarStream := os.Stdin
			if tar2filesopts.tarFile != "" {
				tarStream, err = os.Open(tar2filesopts.tarFile)
				if err != nil {
					return fmt.Errorf("could not open rpm at %s: %v", tar2filesopts.tarFile, err)
				}
				err = rpm.PrefixFilter(tar2filesopts.filePrefix, tar.NewReader(tarStream), args)
				if err != nil {
					return fmt.Errorf("could not convert rpm at %s: %v", tar2filesopts.tarFile, err)
				}
			} else {
				err = rpm.PrefixFilter(tar2filesopts.filePrefix, tar.NewReader(tarStream), args)
				if err != nil {
					return fmt.Errorf("could not convert rpm : %v", err)
				}
			}
			return nil
		},
	}

	tar2filesCmd.Flags().StringVarP(&tar2filesopts.tarFile, "input", "i", "", "location from where to read the tar file (defaults to stdin)")
	tar2filesCmd.Flags().StringVar(&tar2filesopts.filePrefix, "file-prefix", "", "only keep files with this directory prefix")
	return tar2filesCmd
}
