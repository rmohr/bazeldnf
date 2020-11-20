package bazel

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

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
			AddRPMs(file, tt.pkgs)
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
			AddTree("mytree", file, tt.pkgs, []string{"a/x", "b/r", "b/z", "a/g"})
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
	pkg.Checksum = api.Checksum{Text: "1234"}
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
