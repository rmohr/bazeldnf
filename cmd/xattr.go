package main

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rmohr/bazeldnf/pkg/xattr"
)

type xattrOpts struct {
	filePrefix    string
	tarFileInput  string
	tarFileOutput string
	capabilities  map[string]string
	selinuxLabels map[string]string
}

var xattropts = xattrOpts{}

func NewXATTRCmd() *cobra.Command {
	xattrCmd := &cobra.Command{
		Use:   "xattr",
		Short: "Modify xattrs on tar file members",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			capabilityMap := map[string][]string{}
			for file, caps := range xattropts.capabilities {
				split := strings.Split(caps, ":")
				if len(split) > 0 {
					capabilityMap[filepath.Clean(strings.TrimPrefix(file, "/"))] = split
				}
			}
			labelMap := map[string]string{}
			for file, label := range xattropts.selinuxLabels {
				labelMap[filepath.Clean(strings.TrimPrefix(file, "/"))] = label
			}

			streamInput := os.Stdin
			if xattropts.tarFileInput != "" {
				streamInput, err = os.Open(xattropts.tarFileInput)
				if err != nil {
					return err
				}
				defer streamInput.Close()
			}

			streamOutput := os.Stdout
			if xattropts.tarFileOutput != "" {
				streamOutput, err = os.OpenFile(xattropts.tarFileOutput, os.O_WRONLY|os.O_CREATE, os.ModePerm)
				if err != nil {
					return err
				}
			}
			tarWriter := tar.NewWriter(streamOutput)
			defer tarWriter.Close()
			return xattr.Apply(tar.NewReader(streamInput), tarWriter, capabilityMap, labelMap)
		},
	}

	xattrCmd.Flags().StringVarP(&xattropts.tarFileInput, "input", "i", "", "location from where to read the tar file (defaults to stdin)")
	xattrCmd.Flags().StringVarP(&xattropts.tarFileOutput, "output", "o", "", "where to write the file to (defaults to stdout)")
	xattrCmd.Flags().StringToStringVarP(&xattropts.capabilities, "capabilities", "c", map[string]string{}, "capabilities of files (--capabilities=/bin/ls=cap_net_bind_service)")
	xattrCmd.Flags().StringToStringVar(&xattropts.selinuxLabels, "selinux-labels", map[string]string{}, "selinux labels of files (--selinux-labels=/bin/ls=unconfined_u:object_r:default_t:s0)")
	return xattrCmd
}
