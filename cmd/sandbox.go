package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/spf13/cobra"
)

type sandboxOpts struct {
	tar         string
	sandboxRoot string
	name        string
}

var sandboxopts = sandboxOpts{}

func NewSandboxCmd() *cobra.Command {

	sandboxCmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Extract a set of RPMs to a specific location",
		RunE: func(cmd *cobra.Command, objects []string) error {
			rootDir := filepath.Join(sandboxopts.sandboxRoot, "sandbox", sandboxopts.name, "root")
			if err := os.RemoveAll(rootDir); err != nil {
				return fmt.Errorf("failed to do the initial cleanup on the sandbox: %v", err)
			}
			if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create the sandbox base directory: %v", err)
			}
			return rpm.Untar(rootDir, sandboxopts.tar)
		},
	}

	sandboxCmd.Flags().StringVarP(&sandboxopts.tar, "input", "i", "", "Tar file with all dependencies")
	sandboxCmd.Flags().StringVar(&sandboxopts.sandboxRoot, "sandbox", ".bazeldnf", "Root directory of the sandbox")
	sandboxCmd.Flags().StringVarP(&sandboxopts.name, "name", "n", "default", "Sandbox name")
	return sandboxCmd
}
