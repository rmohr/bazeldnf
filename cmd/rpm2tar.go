package main

import (
	"fmt"
	"os"

	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/spf13/cobra"
)

var output string
var input string

func NewRPMCmd() *cobra.Command {
	tarCmd := &cobra.Command{
		Use:   "rpm2tar",
		Short: "convert a rpm to a tar archive",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rpmStream := os.Stdin
			tarStream := os.Stdout
			if input != "-" {
				rpmStream, err = os.Open(input)
				if err != nil {
					return fmt.Errorf("could not open rpm: %v", err)
				}
			}
			if output != "-" {
				tarStream, err = os.Create(output)
				if err != nil {
					return fmt.Errorf("could not create tar: %v", err)
				}
			}
			return rpm.RPMToTar(rpmStream, tarStream)
		},
	}

	tarCmd.PersistentFlags().StringVarP(&output, "output", "o", "-", "location of the resulting tar file (defaults to stdout)")
	tarCmd.PersistentFlags().StringVarP(&input, "input", "i", "-", "location from where to read the rpm file (defaults to stdin)")
	return tarCmd
}
