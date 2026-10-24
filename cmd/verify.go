package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
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
	fromMacro string
	lockfile  string
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
			loadedGPGKeys := map[string]bool{}
			for _, repo := range repos.Repositories {
				if !repo.Disabled && repo.GPGKey != "" && !loadedGPGKeys[repo.GPGKey] {
					loadedGPGKeys[repo.GPGKey] = true
					resp, err := http.Get(repo.GPGKey)
					if err != nil {
						return fmt.Errorf("could not fetch gpgkey %s: %w", repo.GPGKey, err)
					}
					keys, err := openpgp.ReadArmoredKeyRing(resp.Body)
					resp.Body.Close()
					if err != nil {
						return fmt.Errorf("could not load gpgkey %s: %w", repo.GPGKey, err)
					}
					for _, k := range keys {
						keyring = append(keyring, k)
					}
				}
			}

			if verifyopts.lockfile != "" {
				config, err := bazel.LoadLockFile(verifyopts.lockfile)
				if err != nil {
					return fmt.Errorf("failed to open lockfile %s: %w", verifyopts.lockfile, err)
				}
				for _, rpm := range config.RPMs {
					if err := verifyLockfileRPM(config, rpm, keyring); err != nil {
						rpmName := rpm.Name
						if rpmName == "" {
							rpmName = rpm.Id
						}
						return fmt.Errorf("could not verify %s: %w", rpmName, err)
					}
				}
			} else if verifyopts.fromMacro == "" {
				workspace, err := bazel.LoadWorkspace(verifyopts.workspace)
				if err != nil {
					return fmt.Errorf("failed to open workspace %s: %w", verifyopts.workspace, err)
				}
				for _, rpm := range bazel.GetWorkspaceRPMs(workspace) {
					err := verify(rpm, keyring)
					if err != nil {
						return fmt.Errorf("Could not verify %s: %w", rpm.Name(), err)
					}
				}
			} else {
				bzl, defname, err := bazel.ParseMacro(verifyopts.fromMacro)
				if err != nil {
					return fmt.Errorf("failed to parse from-macro expression %q: %w", verifyopts.fromMacro, err)
				}
				bzlfile, err := bazel.LoadBzl(bzl)
				if err != nil {
					return err
				}
				for _, rpm := range bazel.GetBzlfileRPMs(bzlfile, defname) {
					err := verify(rpm, keyring)
					if err != nil {
						return fmt.Errorf("Could not verify %s: %w", rpm.Name(), err)
					}
				}
			}
			return nil
		},
	}

	verifyCmd.Flags().StringArrayVarP(&verifyopts.repofiles, "repofile", "r", []string{"repo.yaml"}, "repository information file (can be specified multiple times)")
	verifyCmd.Flags().StringVarP(&verifyopts.workspace, "workspace", "w", "WORKSPACE", "Bazel workspace file")
	verifyCmd.Flags().StringVarP(&verifyopts.fromMacro, "from-macro", "", "", "Tells bazeldnf to read the RPMs from a macro in the given bzl file instead of the WORKSPACE file. The expected format is: macroFile%defName")
	verifyCmd.Flags().StringVar(&verifyopts.lockfile, "lockfile", "", "lockfile for RPMs")
	return verifyCmd
}

func verify(rpm *bazel.RPMRule, keyring openpgp.EntityList) error {
	return verifyRPM(rpm.Name(), rpm.URLs(), sha256.New, rpm.SHA256(), keyring)
}

func verifyLockfileRPM(config *bazeldnf.Config, rpm *bazeldnf.RPM, keyring openpgp.EntityList) error {
	name := rpm.Name
	if name == "" {
		name = rpm.Id
	}
	urls := rpm.URLs
	if rpm.Repository != "" {
		mirrors, ok := config.Repositories[rpm.Repository]
		if !ok {
			return fmt.Errorf("unknown repository %q", rpm.Repository)
		}
		if len(rpm.URLs) == 0 {
			return fmt.Errorf("no urls defined")
		}
		urls = nil
		for _, mirror := range mirrors {
			u, err := url.Parse(mirror)
			if err != nil {
				return err
			}
			urls = append(urls, u.JoinPath(rpm.URLs[0]).String())
		}
	}
	newHash, expectedHex, err := parseExpectedDigest(rpm.SHA256, rpm.Integrity)
	if err != nil {
		return err
	}
	return verifyRPM(name, urls, newHash, expectedHex, keyring)
}

func parseExpectedDigest(sha256Hex, integrity string) (func() hash.Hash, string, error) {
	if integrity != "" {
		alg, encoded, ok := strings.Cut(integrity, "-")
		if !ok {
			return nil, "", fmt.Errorf("invalid integrity %q", integrity)
		}
		raw, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, "", fmt.Errorf("invalid integrity %q: %w", integrity, err)
		}
		switch alg {
		case "sha256":
			return sha256.New, hex.EncodeToString(raw), nil
		case "sha512":
			return sha512.New, hex.EncodeToString(raw), nil
		case "sha1":
			return sha1.New, hex.EncodeToString(raw), nil
		default:
			return nil, "", fmt.Errorf("unsupported integrity algorithm %q", alg)
		}
	}
	if sha256Hex != "" {
		return sha256.New, strings.ToLower(sha256Hex), nil
	}
	return nil, "", fmt.Errorf("neither sha256 nor integrity set")
}

func verifyRPM(name string, urls []string, newHash func() hash.Hash, expectedHex string, keyring openpgp.EntityList) error {
	// Force a test. If `nil` the verification library just does no GPG check
	if keyring == nil {
		keyring = openpgp.EntityList{}
	}

	log.Infof("Verifying %s", name)
	for _, u := range urls {
		hasher := newHash()
		resp, err := http.Get(u)
		if err != nil {
			log.Warningf("Failed to download %s: %v", name, err)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			resp.Body.Close()
			log.Warningf("Failed to download %s: status %d", name, resp.StatusCode)
			continue
		}
		body := io.TeeReader(resp.Body, hasher)
		_, _, verifyErr := rpmutils.Verify(body, keyring)
		resp.Body.Close()
		var shaErr error
		if expectedHex != toHex(hasher) {
			shaErr = fmt.Errorf("expected checksum %s, but got %s", expectedHex, toHex(hasher))
		}

		if verifyErr != nil && shaErr != nil {
			log.Warningf("Failed to verify %s: %v: %v", name, verifyErr, shaErr)
			continue
		} else if verifyErr != nil {
			return fmt.Errorf("the artifact has the right shasum but is not a RPM: %v", verifyErr)
		} else if shaErr != nil {
			return fmt.Errorf("the artifact is a RPM but not the right one: %v", shaErr)
		}
		return nil
	}
	return fmt.Errorf("could not verify %s", name)
}

func toHex(hasher hash.Hash) string {
	return hex.EncodeToString(hasher.Sum(nil))
}
