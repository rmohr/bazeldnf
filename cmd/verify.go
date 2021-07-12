package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/rmohr/bazeldnf/pkg/bazel"
	"github.com/rmohr/bazeldnf/pkg/repo"
	"github.com/sassoftware/go-rpmutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
)

type VerifyOpts struct {
	repofiles []string
	workspace string
}

var verifyopts = VerifyOpts{}

func NewVerifyCmd() *cobra.Command {

	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "verify RPMs against gpg keys defined in repo.yaml",
		Long:  `verify RPMs against gpg keys defined in repo.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := repo.LoadRepoFiles(verifyopts.repofiles)
			if err != nil {
				return err
			}
			keyring := openpgp.EntityList{}
			for _, repo := range repos.Repositories {
				if !repo.Disabled && repo.GPGKey != "" {
					resp, err := http.Get(repo.GPGKey)
					if err != nil {
						return fmt.Errorf("could not fetch gpgkey %s: %v", repo.GPGKey, err)
					}
					defer resp.Body.Close()
					keys, err := openpgp.ReadArmoredKeyRing(resp.Body)
					if err != nil {
						return fmt.Errorf("could not load gpgkey %s: %v", repo.GPGKey, err)
					}
					for _, k := range keys {
						keyring = append(keyring, k)
					}
				}
			}

			workspace, err := bazel.LoadWorkspace(verifyopts.workspace)
			if err != nil {
				return fmt.Errorf("failed to open workspace %s: %v", verifyopts.workspace, err)
			}
			for _, rpm := range bazel.GetRPMs(workspace) {
				err := verify(rpm, keyring)
				if err != nil {
					return fmt.Errorf("Could not verify %s: %v", rpm.Name(), err)
				}
			}
			return nil
		},
	}

	verifyCmd.Flags().StringArrayVarP(&verifyopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file (can be specified multiple times)")
	verifyCmd.Flags().StringVarP(&verifyopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	return verifyCmd
}

func verify(rpm *bazel.RPMRule, keyring openpgp.EntityList) (err error) {
	// Force a test. If `nil` the verification library just does no GPG check
	if keyring == nil {
		keyring = openpgp.EntityList{}
	}

	log.Infof("Verifying %s", rpm.Name())
	for _, url := range rpm.URLs() {
		sha := sha256.New()
		resp, err := http.Get(url)
		if err != nil {
			log.Warningf("Failed to download %s: %v", rpm.Name(), err)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			log.Warningf("Failed to download %s: %v ", rpm.Name(), fmt.Errorf("status : %v", resp.StatusCode))
			continue
		}
		defer resp.Body.Close()
		body := io.TeeReader(resp.Body, sha)
		_, _, verifyErr := rpmutils.Verify(body, keyring)
		var shaErr error
		if rpm.SHA256() != toHex(sha) {
			shaErr = fmt.Errorf("expected sha256 sum %s, but got %s", rpm.SHA256(), toHex(sha))
		}

		if verifyErr != nil && shaErr != nil {
			log.Warningf("Failed to verify %s: %v: %v", rpm.Name(), verifyErr, shaErr)
			continue
		} else if verifyErr != nil {
			return fmt.Errorf("the artifact has the right shasum but is not a RPM: %v", verifyErr)
		} else if shaErr != nil {
			return fmt.Errorf("the artifact is a RPM but not the right one: %v", shaErr)
		}
		return nil
	}
	return fmt.Errorf("Could not verify %s", rpm.Name())
}

func toHex(hasher hash.Hash) string {
	return hex.EncodeToString(hasher.Sum(nil))
}
