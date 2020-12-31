package main

import (
	"archive/tar"
	"fmt"
	"os"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/order"
	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/spf13/cobra"
)

var output string
var input []string
var symlinks map[string]string
var capabilities map[string]string

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
			cap := map[string][]string{}
			for file, caps := range capabilities {
				split := strings.Split(caps, ":")
				if len(split) > 0 {
					cap["./"+strings.TrimPrefix(file, "/")] = split
				}
			}

			tarWriter := tar.NewWriter(tarStream)
			defer tarWriter.Close()
			if len(input) != 0 {
				directoryTree, err := order.TreeFromRPMs(input)
				if err != nil {
					return err
				}
				for k, v := range symlinks {
					directoryTree.Add(
						[]tar.Header{
							{
								Typeflag: tar.TypeSymlink,
								Name:     k,
								Linkname: v,
							},
						},
					)
				}
				for _, header := range directoryTree.Traverse() {
					err := tarWriter.WriteHeader(&header)
					if err != nil {
						return fmt.Errorf("failed to write header %s: %v", header.Name, err)
					}
				}

				for _, i := range input {
					rpmStream, err = os.Open(i)
					if err != nil {
						return fmt.Errorf("could not open rpm at %s: %v", i, err)
					}
					defer rpmStream.Close()
					err = rpm.RPMToTar(rpmStream, tarWriter, true, cap)
					if err != nil {
						return fmt.Errorf("could not convert rpm at %s: %v", i, err)
					}
				}
			} else {
				err := rpm.RPMToTar(rpmStream, tarWriter, false, cap)
				if err != nil {
					return fmt.Errorf("could not convert rpm : %v", err)
				}
			}
			return nil
		},
	}

	tarCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "location of the resulting tar file (defaults to stdout)")
	tarCmd.PersistentFlags().StringArrayVarP(&input, "input", "i", []string{}, "location from where to read the rpm file (defaults to stdin)")
	tarCmd.Flags().StringToStringVarP(&symlinks, "symlinks", "s", map[string]string{}, "symlinks to add. Relative or absolute.")
	tarCmd.Flags().StringToStringVarP(&capabilities, "capabilties", "c", map[string]string{}, "capabilities of files (-c=/bin/ls=cap_net_bind_service)")
	return tarCmd
}
