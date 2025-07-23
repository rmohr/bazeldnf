package main

import (
	"os"

	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type lockfileOpts struct {
	repofiles  []string
	configname string
	lockfile   string
}

var lockfileopts = lockfileOpts{}

func NewLockFileCmd() *cobra.Command {

	lockfileCmd := &cobra.Command{
		Use:   "lockfile",
		Short: "Manage bazeldnf lock file",
		Long:  `Keep the bazeldnf lock file up to date using a set of dependencies`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, required []string) error {
			repos, err := repo.LoadRepoFiles(lockfileopts.repofiles)
			if err != nil {
				return err
			}

			install, forceIgnored, err := resolve(repos, required)
			if err != nil {
				return err
			}

			logrus.Debugf("install: %v", install)
			logrus.Debugf("forceIgnored: %v", forceIgnored)

			config, err := toConfig(install, forceIgnored, required, os.Args[2:])

			if err != nil {
				return err
			}

			logrus.Info("Writing lockfile.")
			return bazel.WriteLockFile(config, lockfileopts.lockfile)
		},
	}

	addResolveHelperFlags(lockfileCmd)
	repo.AddCacheHelperFlags(lockfileCmd)
	lockfileCmd.Flags().StringArrayVarP(&lockfileopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file. Can be specified multiple times. Will be used by default if no explicit inputs are provided.")
	lockfileCmd.Flags().StringVar(&lockfileopts.configname, "configname", "rpms", "config name to use in lockfile")
	lockfileCmd.Flags().StringVar(&lockfileopts.lockfile, "lockfile", "bazeldnf-lock.json", "lockfile to write to")
	return lockfileCmd
}
