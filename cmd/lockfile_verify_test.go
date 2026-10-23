package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/bazel"
)

func TestNewLockFileHandlerPreservesExistingConfigName(t *testing.T) {
	g := NewGomegaWithT(t)
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "rpms.json")

	// Creating a new lockfile without --configname defaults to "rpms".
	hNew, err := NewLockFileHandler("", filepath.Join(tmpDir, "new.json"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(hNew.config.Name).To(Equal("rpms"))

	// Existing lockfile with custom name ("kubevirt-rpms") preserves its name when --configname is omitted.
	orig := &bazeldnf.Config{
		Name: "kubevirt-rpms",
		RPMs: []*bazeldnf.RPM{
			{Name: "kept-1.0.x86_64", Integrity: "sha256-EjQ=", URLs: []string{"https://example.invalid/kept.rpm"}},
			{Name: "unreferenced-1.0.x86_64", Integrity: "sha256-EjQ=", URLs: []string{"https://example.invalid/drop.rpm"}},
		},
	}
	g.Expect(bazel.WriteLockFile(orig, lockPath)).To(Succeed())

	hExisting, err := NewLockFileHandler("", lockPath)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(hExisting.config.Name).To(Equal("kubevirt-rpms"))

	buildFile, err := build.ParseBuild("BUILD.bazel", []byte(`
rpmtree(
    name = "tree",
    rpms = ["@kubevirt-rpms//kept-1.0.x86_64"],
)
`))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(hExisting.Process(nil, buildFile)).To(Succeed())
	g.Expect(hExisting.config.RPMs).To(HaveLen(1))
	g.Expect(hExisting.config.RPMs[0].Name).To(Equal("kept-1.0.x86_64"))
}

func TestParseExpectedDigest(t *testing.T) {
	sha256Sum := sha256.Sum256([]byte("rpm-bytes"))
	sha512Sum := sha512.Sum512([]byte("rpm-bytes"))
	sha1Sum := sha1.Sum([]byte("rpm-bytes"))

	tests := []struct {
		name      string
		sha256Hex string
		integrity string
		wantHex   string
		wantErr   bool
	}{
		{
			name:      "sri_sha256",
			integrity: "sha256-" + base64.StdEncoding.EncodeToString(sha256Sum[:]),
			wantHex:   hex.EncodeToString(sha256Sum[:]),
		},
		{
			name:      "sri_sha512",
			integrity: "sha512-" + base64.StdEncoding.EncodeToString(sha512Sum[:]),
			wantHex:   hex.EncodeToString(sha512Sum[:]),
		},
		{
			name:      "sri_sha1",
			integrity: "sha1-" + base64.StdEncoding.EncodeToString(sha1Sum[:]),
			wantHex:   hex.EncodeToString(sha1Sum[:]),
		},
		{
			name:      "uppercase_sha256_hex_normalized_to_lowercase",
			sha256Hex: "ABCDEF0123456789",
			wantHex:   "abcdef0123456789",
		},
		{
			name:      "unsupported_sri_algorithm",
			integrity: "md5-" + base64.StdEncoding.EncodeToString([]byte("abc")),
			wantErr:   true,
		},
		{
			name:      "malformed_sri",
			integrity: "sha256_missing_dash",
			wantErr:   true,
		},
		{
			name:    "empty_digests",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			newHash, gotHex, err := parseExpectedDigest(tt.sha256Hex, tt.integrity)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(newHash).ToNot(BeNil())
			g.Expect(gotHex).To(Equal(tt.wantHex))
		})
	}
}

func TestVerifyLockfileRPMValidationErrors(t *testing.T) {
	g := NewGomegaWithT(t)
	cfg := &bazeldnf.Config{
		Repositories: map[string][]string{
			"baseos": {"https://mirror.example/centos"},
		},
	}

	err := verifyLockfileRPM(cfg, &bazeldnf.RPM{
		Id:         "foo-1.0.x86_64",
		Repository: "unknown-repo",
		URLs:       []string{"Packages/foo.rpm"},
		SHA256:     "1234",
	}, nil)
	g.Expect(err).To(MatchError(ContainSubstring(`unknown repository "unknown-repo"`)))

	err = verifyLockfileRPM(cfg, &bazeldnf.RPM{
		Id:         "foo-1.0.x86_64",
		Repository: "baseos",
		SHA256:     "1234",
	}, nil)
	g.Expect(err).To(MatchError(ContainSubstring("no urls defined")))
}
