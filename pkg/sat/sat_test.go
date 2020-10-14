package sat

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazel-dnf/pkg/api"
)

func TestNewResolver(t *testing.T) {
	tests := []struct {
		name     string
		packages []*api.Package
		requires []string
		install  []string
		exclude  []string
		solvable bool
	}{
		{name: "with indirect dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d", "g"}),
			newPkg("testb", "1", []string{"testb", "c"}, []string{}),
			newPkg("testc", "1", []string{"testc", "d"}, []string{}),
			newPkg("testd", "1", []string{"testd", "e", "f", "g"}, []string{"h"}),
			newPkg("teste", "1", []string{"teste", "h"}, []string{}),
		}, requires: []string{
			"testa",
		},
			install:  []string{"testa-0:1", "testc-0:1", "testd-0:1", "teste-0:1"},
			exclude:  []string{"testb-0:1"},
			solvable: true,
		},
		{name: "with circular dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d", "g"}),
			newPkg("testb", "1", []string{"testb", "c"}, []string{}),
			newPkg("testc", "1", []string{"testc", "d"}, []string{}),
			newPkg("testd", "1", []string{"testd", "e", "f", "g"}, []string{"h"}),
			newPkg("teste", "1", []string{"teste", "h"}, []string{"a"}),
		}, requires: []string{
			"testa",
		},
			install: []string{"testa-0:1", "testc-0:1", "testd-0:1", "teste-0:1"},
			exclude: []string{"testb-0:1"},
			solvable: true,
		},
		{name: "with an unresolvable dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d"}),
		}, requires: []string{
			"testa",
		},
			solvable: false,
		},
		{name: "with two sources to choose from", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d"}),
			newPkg("testb", "1", []string{"testb", "d"}, []string{}),
			newPkg("testb", "2", []string{"testb", "d"}, []string{}),
		}, requires: []string{
			"testa",
		},
			install:  []string{"testa-0:1", "testb-0:1"},
			exclude:  []string{"testb-0:2"},
			solvable: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver()
			err := resolver.LoadInvolvedPackages(tt.packages)
			if err != nil {
				t.Fail()
			}
			err = resolver.ConstructRequirements(tt.requires)
			if err != nil {
				fmt.Println(err)
				t.Fail()
			}
			install, exclude, err := resolver.Resolve()
			g := NewGomegaWithT(t)
			if tt.solvable {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(install).To(ConsistOf(strToPkg(tt.install, tt.packages)))
			g.Expect(exclude).To(ConsistOf(strToPkg(tt.exclude, tt.packages)))
		})
	}
}

func newPkg(name string, version string, provides []string, requires []string) *api.Package {
	pkg := &api.Package{}
	pkg.Name = name
	pkg.Version = api.Version{Ver: version}
	for _, req := range requires {
		pkg.Format.Requires.Entries = append(pkg.Format.Requires.Entries, api.Entry{Name: req})
	}
	pkg.Format.Provides.Entries = append(pkg.Format.Provides.Entries, api.Entry{
		Name:  name,
		Flags: "EQ",
		Ver:   version,
	})
	for _, req := range provides {
		pkg.Format.Provides.Entries = append(pkg.Format.Provides.Entries, api.Entry{Name: req})
	}

	return pkg
}

func strToPkg(wanted []string, given []*api.Package) (resolved []*api.Package) {
	m := map[string]*api.Package{}
	for _, p := range given {
		m[p.String()] = p
	}
	for _, w := range wanted {
		resolved = append(resolved, m[w])
	}
	return resolved
}