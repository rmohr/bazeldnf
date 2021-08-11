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

type rpm2tarOpts struct {
	output        string
	input         []string
	symlinks      map[string]string
	capabilities  map[string]string
	selinuxLabels map[string]string
}

var rpm2taropts = rpm2tarOpts{}

func NewRpm2TarCmd() *cobra.Command {
	rpm2tarCmd := &cobra.Command{
		Use:   "rpm2tar",
		Short: "convert a rpm to a tar archive",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rpmStream := os.Stdin
			tarStream := os.Stdout
			if rpm2taropts.output != "" {
				tarStream, err = os.Create(rpm2taropts.output)
				if err != nil {
					return fmt.Errorf("could not create tar: %v", err)
				}
			}
			cap := map[string][]string{}
			for file, caps := range rpm2taropts.capabilities {
				split := strings.Split(caps, ":")
				if len(split) > 0 {
					cap["./"+strings.TrimPrefix(file, "/")] = split
				}
			}

			tarWriter := tar.NewWriter(tarStream)
			defer tarWriter.Close()
			if len(rpm2taropts.input) != 0 {
				directoryTree, err := order.TreeFromRPMs(rpm2taropts.input)
				if err != nil {
					return err
				}
				for k, v := range rpm2taropts.symlinks {
					// If an absolute path is given let's add a `.` in front. This is
					// not strictly necessary but adds a more correct tar path
					// which aligns with the usual rpm entries which start with `./`
					if strings.HasPrefix(k, "/") {
						k = "." + k
					}
					directoryTree.Add(
						[]tar.Header{
							{
								Typeflag: tar.TypeSymlink,
								Name:     k,
								Linkname: v,
								Mode:     0777,
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

				for _, i := range rpm2taropts.input {
					rpmStream, err = os.Open(i)
					if err != nil {
						return fmt.Errorf("could not open rpm at %s: %v", i, err)
					}
					defer rpmStream.Close()
					err = rpm.RPMToTar(rpmStream, tarWriter, true, cap, rpm2taropts.selinuxLabels)
					if err != nil {
						return fmt.Errorf("could not convert rpm at %s: %v", i, err)
					}
				}
			} else {
				err := rpm.RPMToTar(rpmStream, tarWriter, false, cap, rpm2taropts.selinuxLabels)
				if err != nil {
					return fmt.Errorf("could not convert rpm : %v", err)
				}
			}
			return nil
		},
	}

	rpm2tarCmd.Flags().StringVarP(&rpm2taropts.output, "output", "o", "", "location of the resulting tar file (defaults to stdout)")
	rpm2tarCmd.Flags().StringArrayVarP(&rpm2taropts.input, "input", "i", []string{}, "location from where to read the rpm file (defaults to stdin)")
	rpm2tarCmd.Flags().StringToStringVarP(&rpm2taropts.symlinks, "symlinks", "s", map[string]string{}, "symlinks to add. Relative or absolute.")
	rpm2tarCmd.Flags().StringToStringVarP(&rpm2taropts.capabilities, "capabilities", "c", map[string]string{}, "capabilities of files (--capabilities=/bin/ls=cap_net_bind_service)")
	rpm2tarCmd.Flags().StringToStringVar(&rpm2taropts.selinuxLabels, "selinux-labels", map[string]string{}, "selinux labels of files (--selinux-labels=/bin/ls=unconfined_u:object_r:default_t:s0)")
	// deprecated options
	rpm2tarCmd.Flags().StringToStringVar(&rpm2taropts.capabilities, "capabilties", map[string]string{}, "capabilities of files (-c=/bin/ls=cap_net_bind_service)")
	rpm2tarCmd.Flags().MarkDeprecated("capabilties", "use --capabilities instead")
	rpm2tarCmd.Flags().MarkShorthandDeprecated("capabilities", "use --capabilities instead")
	rpm2tarCmd.Flags().MarkShorthandDeprecated("symlinks", "use --symlinks instead")
	return rpm2tarCmd
}
