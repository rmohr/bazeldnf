package bazel

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
)


func TestPruneRPMs(t *testing.T) {
	tests := []struct {
		name string
		orig string
		expected string
		pkgs []*api.Package
	}{
		{
			name: "should remove rpm entries",
			orig: "testdata/WORKSPACE",
			expected: "testdata/WORKSPACE.norpm",
		},
		{
			name: "should replace rpm entries",
			orig: "testdata/WORKSPACE",
			expected: "testdata/WORKSPACE.pkgs",
			pkgs: []*api.Package{
				newPkg("a", "1.2.3"),
				newPkg("a", "2.3.4"),
				newPkg("b", "2.3.4"),
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
			PruneRPMs(file)
			AddRPMS(file, tt.pkgs)
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


func newPkg(name string, version string) *api.Package {
	pkg := &api.Package{}
	pkg.Name = name
	pkg.Version = api.Version{Ver: version}
	return pkg
}
