package main

import (
	"errors"
	"slices"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

func TestBaseCase(t *testing.T) {
	g := NewGomegaWithT(t)

	expected := &bazeldnf.Config{
		CommandLineArguments: []string{},
		Name:                 "",
		Repositories:         map[string][]string{},
		RPMs:                 []*bazeldnf.RPM{},
		Targets:              []string{},
		ForceIgnored:         []string{},
	}
	cfg, err := toConfig([]*api.Package{}, []*api.Package{}, []string{}, []string{})

	g.Expect(err).Should(BeNil())
	g.Expect(cfg).Should(Equal(expected))
}

func TestSimpleInputs(t *testing.T) {
	g := NewGomegaWithT(t)

	ignored := []*api.Package{
		&api.Package{Name: "package0"},
		&api.Package{Name: "package1"},
	}
	targets := []string{"foo", "bar", "baz"}
	commandline := []string{"baf", "bam"}

	expected := &bazeldnf.Config{
		CommandLineArguments: commandline,
		Name:                 "",
		Repositories:         map[string][]string{},
		RPMs:                 []*bazeldnf.RPM{},
		Targets:              targets,
		ForceIgnored:         []string{"package0", "package1"},
	}
	cfg, err := toConfig([]*api.Package{}, ignored, targets, commandline)

	g.Expect(err).Should(BeNil())
	g.Expect(cfg).Should(Equal(expected))
}

func TestMissingProvider(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg, err := toConfig(
		[]*api.Package{
			newPackageWithDeps("parent", "somedep"),
		},
		[]*api.Package{},
		[]string{"parent"},
		[]string{},
	)

	g.Expect(err).Should(Equal(errors.New("could not find provider for somedep")))
	g.Expect(cfg).Should(BeNil())
}

func newPackage(name, checksum, url, repository string, mirrors []string) *api.Package {
	return &api.Package{
		Name:     name,
		Checksum: api.Checksum{Text: checksum, Type: "sha256"},
		Location: api.Location{Href: url},
		Repository: &bazeldnf.Repository{
			Name:    repository,
			Mirrors: mirrors,
		},
	}
}

func newSimplePackage(name string) *api.Package {
	return newPackage(name, "", "", "repository", []string{})
}

func newPackageWithDeps(name string, deps ...string) *api.Package {
	p := newSimplePackage(name)
	entries := []api.Entry{}
	for _, dep := range deps {
		entries = append(entries, api.Entry{Name: dep})
	}
	p.Format.Requires.Entries = entries
	p.Format.Provides.Entries = []api.Entry{api.Entry{Name: name}}

	return p
}

func newPackageWithFiles(name string, files ...string) *api.Package {
	p := newSimplePackage(name)
	entries := []api.ProvidedFile{}
	for _, file := range files {
		entries = append(entries, api.ProvidedFile{Text: file})
	}
	p.Format.Files = entries
	return p
}

func newSimpleRPM(name string, deps ...string) *bazeldnf.RPM {
	d := []string{}
	if len(deps) > 0 {
		d = deps
	}

	return &bazeldnf.RPM{
		Name:         name,
		URLs:         []string{""},
		Integrity:    "sha256-",
		Repository:   "repository",
		Dependencies: d,
	}
}

type testCaseConfiguration struct {
	name                 string
	installed, ignored   []*api.Package
	expectedRepositories map[string][]string
	expectedRPMs         []*bazeldnf.RPM
	requested            []string
}

func testCase(config testCaseConfiguration, t *testing.T) {
	t.Run(config.name, func(t *testing.T) {
		g := NewGomegaWithT(t)

		forceIgnored := []string{}
		for _, pkg := range config.ignored {
			forceIgnored = append(forceIgnored, pkg.Name)
		}
		slices.Sort(forceIgnored)

		expected := &bazeldnf.Config{
			CommandLineArguments: []string{},
			Name:                 "",
			Repositories:         config.expectedRepositories,
			RPMs:                 config.expectedRPMs,
			Targets:              config.requested,
			ForceIgnored:         forceIgnored,
		}

		cfg, err := toConfig(
			config.installed,
			config.ignored,
			config.requested,
			[]string{},
		)

		g.Expect(err).Should(BeNil())
		g.Expect(cfg).Should(Equal(expected))
	})
}

func TestOneInstalled(t *testing.T) {
	testCase(testCaseConfiguration{
		name: "one installed",
		installed: []*api.Package{
			newPackage(
				"package0",
				"f87b49c517aac9eb4890a4b5005bcc4a586748f2760ea1106382f3897129a60e",
				"urlforrpm",
				"repository",
				[]string{"mirror0", "mirror1"},
			),
		},
		ignored: []*api.Package{},
		expectedRepositories: map[string][]string{
			"repository": []string{"mirror0", "mirror1"},
		},
		expectedRPMs: []*bazeldnf.RPM{
			&bazeldnf.RPM{
				Name:         "package0",
				Integrity:    "sha256-+HtJxReqyetIkKS1AFvMSlhnSPJ2DqEQY4LziXEppg4=",
				URLs:         []string{"urlforrpm"},
				Repository:   "repository",
				Dependencies: []string{},
			},
		},
	}, t)
}

func TestConfigTransform(t *testing.T) {
	tests := []struct {
		name               string
		installed, ignored []*api.Package

		expectedRepositories map[string][]string
		expectedRPMs         []*bazeldnf.RPM
		requested            []string
	}{
		{
			name: "one installed",
			installed: []*api.Package{
				newPackage(
					"package0",
					"f87b49c517aac9eb4890a4b5005bcc4a586748f2760ea1106382f3897129a60e",
					"urlforrpm",
					"repository",
					[]string{"mirror0", "mirror1"},
				),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{"mirror0", "mirror1"},
			},
			expectedRPMs: []*bazeldnf.RPM{
				&bazeldnf.RPM{
					Name:         "package0",
					Integrity:    "sha256-+HtJxReqyetIkKS1AFvMSlhnSPJ2DqEQY4LziXEppg4=",
					URLs:         []string{"urlforrpm"},
					Repository:   "repository",
					Dependencies: []string{},
				},
			},
		},
		{
			name: "two installed",
			installed: []*api.Package{
				newPackage(
					"package0",
					"f87b49c517aac9eb4890a4b5005bcc4a586748f2760ea1106382f3897129a60e",
					"urlforrpm",
					"repository",
					[]string{"mirror0", "mirror1"},
				),
				newPackage(
					"package1",
					"9146a02ed928ffca6ef0f1241d2d86e4e998e6f70aae875754601fda54951fbd",
					"urlforrpm0",
					"repository0",
					[]string{},
				),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository":  []string{"mirror0", "mirror1"},
				"repository0": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				&bazeldnf.RPM{
					Name:         "package0",
					Integrity:    "sha256-+HtJxReqyetIkKS1AFvMSlhnSPJ2DqEQY4LziXEppg4=",
					URLs:         []string{"urlforrpm"},
					Repository:   "repository",
					Dependencies: []string{},
				},
				&bazeldnf.RPM{
					Name:         "package1",
					Integrity:    "sha256-kUagLtko/8pu8PEkHS2G5OmY5vcKrodXVGAf2lSVH70=",
					URLs:         []string{"urlforrpm0"},
					Repository:   "repository0",
					Dependencies: []string{},
				},
			},
		},
		{
			name: "two installed repo overlap",
			installed: []*api.Package{
				newPackage(
					"package0",
					"f87b49c517aac9eb4890a4b5005bcc4a586748f2760ea1106382f3897129a60e",
					"urlforrpm",
					"repository",
					[]string{"mirror0", "mirror1"},
				),
				newPackage(
					"package1",
					"9146a02ed928ffca6ef0f1241d2d86e4e998e6f70aae875754601fda54951fbd",
					"urlforrpm0",
					"repository",
					[]string{},
				),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				&bazeldnf.RPM{
					Name:         "package0",
					Integrity:    "sha256-+HtJxReqyetIkKS1AFvMSlhnSPJ2DqEQY4LziXEppg4=",
					URLs:         []string{"urlforrpm"},
					Repository:   "repository",
					Dependencies: []string{},
				},
				&bazeldnf.RPM{
					Name:         "package1",
					Integrity:    "sha256-kUagLtko/8pu8PEkHS2G5OmY5vcKrodXVGAf2lSVH70=",
					URLs:         []string{"urlforrpm0"},
					Repository:   "repository",
					Dependencies: []string{},
				},
			},
		},
		{
			name: "two installed out of order",
			installed: []*api.Package{
				newSimplePackage("package2"),
				newSimplePackage("package1"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1"),
				newSimpleRPM("package2"),
			},
		},
		{
			name: "two installed dep between",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2"),
				newPackageWithDeps("package2"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2"),
				newSimpleRPM("package2"),
			},
			requested: []string{"package1"},
		},
		{
			name: "three installed dep from first",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2", "package3"),
				newPackageWithDeps("package2"),
				newPackageWithDeps("package3"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2", "package3"),
				newSimpleRPM("package2"),
				newSimpleRPM("package3"),
			},
			requested: []string{"package1"},
		},
		{
			name: "three installed dep from first sort deps",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package3", "package2"),
				newPackageWithDeps("package2"),
				newPackageWithDeps("package3"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2", "package3"),
				newSimpleRPM("package2"),
				newSimpleRPM("package3"),
			},
			requested: []string{"package1"},
		},
		{
			name: "three installed dep transitive",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2"),
				newPackageWithDeps("package2", "package3"),
				newPackageWithDeps("package3"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2", "package3"),
				newSimpleRPM("package2"),
				newSimpleRPM("package3"),
			},
			requested: []string{"package1"},
		},
		{
			name: "three installed dep overlap",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package3"),
				newPackageWithDeps("package2", "package3"),
				newPackageWithDeps("package3"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package3"),
				newSimpleRPM("package2", "package3"),
				newSimpleRPM("package3"),
			},
			requested: []string{"package1", "package2"},
		},
		{
			name: "two installed require ignored",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2"),
			},
			ignored: []*api.Package{
				newPackageWithDeps("package2"),
			},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1"),
			},
		},
		{
			name: "depends on self",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package1"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1"),
			},
		},
		{
			name:      "sort ignored",
			installed: []*api.Package{},
			ignored: []*api.Package{
				newPackageWithDeps("package2"),
				newPackageWithDeps("package1"),
			},
			expectedRepositories: map[string][]string{},
			expectedRPMs:         []*bazeldnf.RPM{},
		},
		{
			name: "file based deps",
			installed: []*api.Package{
				newPackageWithDeps("package1", "somefile"),
				newPackageWithFiles("package2", "somefile"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2"),
				newSimpleRPM("package2"),
			},
			requested: []string{"package1"},
		},
		{
			name: "file based deps ignored provider",
			installed: []*api.Package{
				newPackageWithDeps("package1", "somefile"),
			},
			ignored: []*api.Package{
				newPackageWithFiles("package2", "somefile"),
			},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1"),
			},
		},
		{
			name: "circular dependencies",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2"),
				newPackageWithDeps("package2", "package4"),
				newPackageWithDeps("package3", "package2"),
				newPackageWithDeps("package4"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2", "package4"),
				newSimpleRPM("package2", "package4"),
				newSimpleRPM("package3", "package2", "package4"),
				newSimpleRPM("package4"),
			},
			requested: []string{"package1", "package2", "package3"},
		},
		{
			name: "transitive circular dependencies",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2"),
				newPackageWithDeps("package2", "package4"),
				newPackageWithDeps("package3", "package2"),
				newPackageWithDeps("package4"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2", "package4"),
				newSimpleRPM("package2"),
				newSimpleRPM("package3", "package2", "package4"),
				newSimpleRPM("package4"),
			},
			requested: []string{"package1", "package3"},
		},
		{
			name: "transitive circular dependencies with more than one cycle",
			installed: []*api.Package{
				newPackageWithDeps("package1", "package2"),
				newPackageWithDeps("package2", "package3"),
				newPackageWithDeps("package3", "package1"),
			},
			ignored: []*api.Package{},
			expectedRepositories: map[string][]string{
				"repository": []string{},
			},
			expectedRPMs: []*bazeldnf.RPM{
				newSimpleRPM("package1", "package2", "package3"),
				newSimpleRPM("package2", "package1", "package3"),
				newSimpleRPM("package3", "package1", "package2"),
			},
			requested: []string{"package1", "package2", "package3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			forceIgnored := []string{}
			for _, pkg := range tt.ignored {
				forceIgnored = append(forceIgnored, pkg.Name)
			}
			slices.Sort(forceIgnored)

			expected := &bazeldnf.Config{
				CommandLineArguments: []string{},
				Name:                 "",
				Repositories:         tt.expectedRepositories,
				RPMs:                 tt.expectedRPMs,
				Targets:              tt.requested,
				ForceIgnored:         forceIgnored,
			}

			cfg, err := toConfig(
				tt.installed,
				tt.ignored,
				tt.requested,
				[]string{},
			)

			g.Expect(err).Should(BeNil())
			g.Expect(cfg).Should(Equal(expected))
		})
	}
}
