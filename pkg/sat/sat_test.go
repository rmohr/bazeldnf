package sat

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazel-dnf/pkg/api"
)

func Test(t *testing.T) {
	t.Run("should resolve bash", func(t *testing.T) {
		g := NewGomegaWithT(t)
		f, err := os.Open("../../testdata/bash-fc31.xml")
		g.Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		repo := &api.Repository{}
		err = xml.NewDecoder(f).Decode(repo)
		g.Expect(err).ToNot(HaveOccurred())

		resolver := NewResolver(false)
		packages := []*api.Package{}
		for i, _ := range repo.Packages {
			packages = append(packages, &repo.Packages[i])
		}
		err = resolver.LoadInvolvedPackages(packages)
		g.Expect(err).ToNot(HaveOccurred())
		err = resolver.ConstructRequirements([]string{"bash", "fedora-release-server", "glibc-langpack-en"})
		g.Expect(err).ToNot(HaveOccurred())
		install, _, err := resolver.Resolve()
		g.Expect(pkgToString(install)).To(ConsistOf(
			"libgcc-0:10.2.1-1.fc32",
			"fedora-gpg-keys-0:32-6",
			"glibc-0:2.31-4.fc32",
			"glibc-langpack-en-0:2.31-4.fc32",
			"fedora-release-common-0:32-3",
			"glibc-common-0:2.31-4.fc32",
			"ncurses-base-0:6.1-15.20191109.fc32",
			"ncurses-libs-0:6.1-15.20191109.fc32",
			"fedora-release-server-0:32-3",
			"tzdata-0:2020a-1.fc32",
			"setup-0:2.13.6-2.fc32",
			"basesystem-0:11-9.fc32",
			"bash-0:5.0.17-1.fc32",
			"filesystem-0:3.14-2.fc32",
			"fedora-repos-0:32-6",
		))
		g.Expect(err).ToNot(HaveOccurred())
	})
}

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
			install:  []string{"testa-0:1", "testc-0:1", "testd-0:1", "teste-0:1"},
			exclude:  []string{"testb-0:1"},
			solvable: true,
		},
		{name: "with an unresolvable dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d"}),
		}, requires: []string{
			"testa",
		},
			solvable: false,
		},
		{name: "with two sources to choose from, should use the newer one", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d"}),
			newPkg("testb", "1", []string{"testb", "d"}, []string{}),
			newPkg("testb", "2", []string{"testb", "d"}, []string{}),
		}, requires: []string{
			"testa",
		},
			install:  []string{"testa-0:1", "testb-0:2"},
			exclude:  []string{},
			solvable: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(false)
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
			g.Expect(pkgToString(install)).To(ConsistOf(tt.install))
			g.Expect(pkgToString(exclude)).To(ConsistOf(tt.exclude))
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

func pkgToString(given []*api.Package) (resolved []string) {
	for _, p := range given {
		resolved = append(resolved, p.String())
	}
	return
}
