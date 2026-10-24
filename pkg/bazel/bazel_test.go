package bazel

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

func TestWorkspaceWithRPMs(t *testing.T) {
	tests := []struct {
		name     string
		orig     string
		expected string
		pkgs     []*api.Package
	}{
		{
			name:     "should replace rpm entries",
			orig:     "testdata/WORKSPACE",
			expected: "testdata/WORKSPACE.pkgs",
			pkgs: []*api.Package{
				newPkg("a", "1.2.3", repo("a", []string{"a", "b", "c"})),
				newPkg("a", "2.3.4", repo("b", []string{"e", "f", "g"})),
				newPkg("b", "2.3.4", repo("a", []string{"a", "b", "c"})),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			tmpFile, err := ioutil.TempFile(os.TempDir(), "WORKSPACE")
			g.Expect(err).ToNot(HaveOccurred())
			defer os.Remove(tmpFile.Name())
			file, err := LoadWorkspace(tt.orig)
			g.Expect(err).ToNot(HaveOccurred())
			AddWorkspaceRPMs(file, tt.pkgs)
			err = WriteWorkspace(false, file, tmpFile.Name())
			g.Expect(err).ToNot(HaveOccurred())

			current, err := ioutil.ReadFile(tmpFile.Name())
			g.Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile(tt.expected)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(current)).To(Equal(string(expected)))
		})
	}
}

func TestBzlfileWithRPMs(t *testing.T) {
	tests := []struct {
		name     string
		orig     string
		expected string
		pkgs     []*api.Package
	}{
		{
			name:     "should replace rpm entries",
			orig:     "testdata/orig.bzl",
			expected: "testdata/pkgs.bzl",
			pkgs: []*api.Package{
				newPkg("a", "1.2.3", repo("a", []string{"a", "b", "c"})),
				newPkg("a", "2.3.4", repo("b", []string{"e", "f", "g"})),
				newPkg("b", "2.3.4", repo("a", []string{"a", "b", "c"})),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			tmpFile, err := ioutil.TempFile(os.TempDir(), "tmp.bzl")
			g.Expect(err).ToNot(HaveOccurred())
			defer os.Remove(tmpFile.Name())
			file, err := LoadWorkspace(tt.orig)
			g.Expect(err).ToNot(HaveOccurred())
			AddBzlfileRPMs(file, "rpms", tt.pkgs)
			err = WriteWorkspace(false, file, tmpFile.Name())
			g.Expect(err).ToNot(HaveOccurred())

			current, err := ioutil.ReadFile(tmpFile.Name())
			g.Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile(tt.expected)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(current)).To(Equal(string(expected)))
		})
	}
}

func TestBuildfileWithRPMs(t *testing.T) {
	tests := []struct {
		name     string
		orig     string
		expected string
		pkgs     []*api.Package
	}{
		{
			name:     "should replace rpm entries",
			orig:     "testdata/BUILD.bazel.test",
			expected: "testdata/BUILD.bazel.result",
			pkgs: []*api.Package{
				newPkg("a", "1.2.3", repo("a", []string{"a", "b", "c"})),
				newPkg("a", "2.3.4", repo("b", []string{"e", "f", "g"})),
				newPkg("b", "2.3.4", repo("a", []string{"a", "b", "c"})),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			tmpFile, err := ioutil.TempFile(os.TempDir(), "BUILD")
			g.Expect(err).ToNot(HaveOccurred())
			defer os.Remove(tmpFile.Name())
			file, err := LoadBuild(tt.orig)
			g.Expect(err).ToNot(HaveOccurred())
			AddTree("mytree", "", file, tt.pkgs, false)
			err = WriteBuild(false, file, tmpFile.Name())
			g.Expect(err).ToNot(HaveOccurred())

			current, err := ioutil.ReadFile(tmpFile.Name())
			g.Expect(err).ToNot(HaveOccurred())
			fmt.Println(string(current))
			expected, err := ioutil.ReadFile(tt.expected)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(current)).To(Equal(string(expected)))
		})
	}
}

func newPkg(name string, version string, repository *bazeldnf.Repository) *api.Package {
	pkg := &api.Package{}
	pkg.Name = name
	pkg.Arch = "myarch"
	pkg.Checksum = api.Checksum{Type: "sha256", Text: "1234"}
	pkg.Version = api.Version{Ver: version}
	pkg.Repository = repository
	pkg.Location = api.Location{Href: "something/" + name}
	return pkg
}

func repo(name string, urls []string) *bazeldnf.Repository {
	return &bazeldnf.Repository{
		Name:    name,
		Mirrors: urls,
	}
}

func TestConfigRPMsAndPrune(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg := &bazeldnf.Config{
		Name: "kubevirt-rpms",
		RPMs: []*bazeldnf.RPM{
			{
				Name:   "a-0__1.2.3.myarch",
				SHA256: "1234",
				URLs:   []string{"https://mirror.orig/something/a", "https://storage.googleapis.com/builddeps/1234"},
			},
			{
				Name:      "stale-0__0.1.0.myarch",
				Integrity: "sha256-EjQ=",
				URLs:      []string{"https://mirror.orig/something/stale"},
			},
		},
	}

	pkgAUnchanged := newPkg("a", "1.2.3", repo("baseos", []string{"https://new-mirror.example"}))
	pkgBUpdated := newPkg("b", "2.0.0", repo("appstream", []string{"https://mirror.example"}))

	g.Expect(AddConfigRPMs(cfg, []*api.Package{pkgBUpdated, pkgAUnchanged})).To(Succeed())
	g.Expect(cfg.RPMs).To(HaveLen(3))
	// Sorted by Name: a, b, stale; 'a' preserves existing URLs (including GCS mirror) because checksum matched.
	g.Expect(cfg.RPMs[0].Name).To(Equal("a-0__1.2.3.myarch"))
	g.Expect(cfg.RPMs[0].Integrity).To(Equal("sha256-EjQ="))
	g.Expect(cfg.RPMs[0].SHA256).To(BeEmpty())
	g.Expect(cfg.RPMs[0].URLs).To(Equal([]string{
		"https://mirror.orig/something/a",
		"https://storage.googleapis.com/builddeps/1234",
	}))
	g.Expect(cfg.RPMs[1].Name).To(Equal("b-0__2.0.0.myarch"))
	g.Expect(cfg.RPMs[1].URLs).To(Equal([]string{"https://mirror.example/something/b"}))

	// When checksum changes on 'a', URLs are refreshed from repository mirrors and Repository is cleared.
	cfg.RPMs[0].Repository = "baseos"
	pkgAChanged := newPkg("a", "1.2.3", repo("baseos", []string{"https://new-mirror.example"}))
	pkgAChanged.Checksum = api.Checksum{Type: "sha256", Text: "5678"}
	g.Expect(AddConfigRPMs(cfg, []*api.Package{pkgAChanged})).To(Succeed())
	g.Expect(cfg.RPMs[0].URLs).To(Equal([]string{"https://new-mirror.example/something/a"}))
	g.Expect(cfg.RPMs[0].Repository).To(BeEmpty())

	// Prune keeps only RPMs referenced in rpmtree rules.
	buildFile, err := build.ParseBuild("BUILD.bazel", []byte(`
rpmtree(
    name = "tree1",
    rpms = [
        "@kubevirt-rpms//a-0__1.2.3.myarch",
        "@kubevirt-rpms//b-0__2.0.0.myarch",
    ],
)
`))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(PruneConfigRPMs(buildFile, "", cfg)).To(Succeed())
	g.Expect(cfg.RPMs).To(HaveLen(2))
	g.Expect(cfg.RPMs[0].Name).To(Equal("a-0__1.2.3.myarch"))
	g.Expect(cfg.RPMs[1].Name).To(Equal("b-0__2.0.0.myarch"))

	// Prune errors when both configname and config.Name are empty.
	g.Expect(PruneConfigRPMs(buildFile, "", &bazeldnf.Config{})).To(MatchError(ContainSubstring("must not be empty")))

	lockPath := filepath.Join(t.TempDir(), "rpms.json")
	g.Expect(WriteLockFile(cfg, lockPath)).To(Succeed())
	loaded, err := LoadLockFile(lockPath)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(loaded).To(Equal(cfg))
}
