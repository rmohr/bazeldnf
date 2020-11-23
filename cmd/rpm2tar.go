package main

import (
	"archive/tar"
	"fmt"
	"os"

	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/spf13/cobra"
)

var output string
var input []string

func NewRPMCmd() *cobra.Command {
	tarCmd := &cobra.Command{
		Use:   "rpm2tar",
		Short: "convert a rpm to a tar archive",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rpmStream := os.Stdin
			tarStream := os.Stdout
			if output != "" {
				tarStream, err = os.Create(output)
				if err != nil {
					return fmt.Errorf("could not create tar: %v", err)
				}
			}
			tarWriter := tar.NewWriter(tarStream)
			defer tarWriter.Close()
			if len(input) != 0 {
				for _, i := range input {
					rpmStream, err = os.Open(i)
					if err != nil {
						return fmt.Errorf("could not open rpm at %s: %v", i, err)
					}
					err = rpm.RPMToTar(rpmStream, tarWriter)
					if err != nil {
						return fmt.Errorf("could not convert rpm at %s: %v", i, err)
					}
				}
			} else {
				err := rpm.RPMToTar(rpmStream, tarWriter)
				if err != nil {
					return fmt.Errorf("could not convert rpm : %v", err)
				}
			}
			return nil
		},
	}

	tarCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "location of the resulting tar file (defaults to stdout)")
	tarCmd.PersistentFlags().StringArrayVarP(&input, "input", "i", []string{}, "location from where to read the rpm file (defaults to stdin)")
	return tarCmd
}
