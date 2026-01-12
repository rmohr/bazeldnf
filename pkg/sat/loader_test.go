package sat

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/crillab/gophersat/bf"
	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

func expectedVars(g *WithT, m *Model, vars ...string) {
	g.Expect(m.vars).To(HaveLen(len(vars)))

	for n, v := range vars {
		k := fmt.Sprintf("x%d", n+1)
		g.Expect(m.vars[k].String()).To(Equal(v))
	}
}

func expectedPackages(g *WithT, m *Model, pkgs map[string][]string) {
	g.Expect(m.packages).To(HaveLen(len(pkgs)))

	for k, versions := range pkgs {
		g.Expect(m.packages[k]).To(HaveLen(len(versions)))

		for n, version := range versions {
			pkg := m.packages[k][n].Package
			g.Expect(pkg.Version.String()).To(Equal(version))
		}
	}
}

func expectedBest(g *WithT, m *Model, pkgVersions map[string]string) {
	g.Expect(m.bestPackages).To(HaveLen(len(pkgVersions)))

	for k, version := range pkgVersions {
		g.Expect(m.bestPackages[k].Version.String()).To(Equal(version))
	}
}

func expectedIgnores(g *WithT, m *Model, pkgKeys ...api.PackageKey) {
	g.Expect(m.forceIgnoreWithDependencies).To(HaveLen(len(pkgKeys)))

	for _, key := range pkgKeys {
		g.Expect(m.forceIgnoreWithDependencies).To(HaveKey(key))
	}
}

func expectedAnds(g *WithT, m *Model, ands ...bf.Formula) {
	n := len(m.vars)
	permutations := int(math.Pow(2, float64(n)))

	for i := 0; i < permutations; i++ {
		c := make(map[string]bool)
		for j := 0; j < n; j++ {
			v := fmt.Sprintf("x%d", n-j)
			c[v] = (i>>j)&1 == 1
		}

		g.Expect(bf.And(m.ands...).Eval(c)).To(Equal(bf.And(ands...).Eval(c)))
	}
}

func newVersion(versionStr string) api.Version {
	// versionStr is formatted as epoch:version-release where epoch
	// defaults to 0 if not provided and "-release" is optional
	parts := strings.SplitN(versionStr, ":", 2)

	epoch := "0"
	if len(parts) == 2 {
		epoch = parts[0]
	}

	ver := parts[len(parts)-1]
	rel := ""

	if lastHyphen := strings.LastIndex(ver, "-"); lastHyphen != -1 {
		rel = ver[lastHyphen+1:]
		ver = ver[:lastHyphen]
	}

	return api.Version{
		Epoch: epoch,
		Ver:   ver,
		Rel:   rel,
	}
}

func newEntry(entryStr string) api.Entry {
	fields := strings.Fields(entryStr)

	entry := api.Entry{Name: fields[0]}
	if len(fields) > 2 {
		entry.Flags = fields[1]
		ver := newVersion(fields[2])
		entry.Epoch, entry.Ver, entry.Rel = ver.Epoch, ver.Ver, ver.Rel
	}

	return entry
}

func toEntries(entryStrs []string) (entries []api.Entry) {
	for _, entryStr := range entryStrs {
		entries = append(entries, newEntry(entryStr))
	}
	return entries
}

func newPackage(
	name, versionStr string,
	requires, provides, conflicts, files []string,
) *api.Package {
	pkg := &api.Package{
		Name:       name,
		Repository: &bazeldnf.Repository{},
		Version:    newVersion(versionStr),
	}

	// Every package provides itself with a specific version
	pkg.Format.Provides.Entries = append(pkg.Format.Provides.Entries, newEntry(
		name+" EQ "+versionStr,
	))

	pkg.Format.Requires.Entries = append(pkg.Format.Requires.Entries, toEntries(requires)...)
	pkg.Format.Provides.Entries = append(pkg.Format.Provides.Entries, toEntries(provides)...)
	pkg.Format.Conflicts.Entries = append(pkg.Format.Conflicts.Entries, toEntries(conflicts)...)
	for _, f := range files {
		pkg.Format.Files = append(pkg.Format.Files, api.ProvidedFile{Text: f})
	}
	return pkg
}

func newSimplePackage(name, versionStr string) *api.Package {
	return newPackage(name, versionStr, nil, nil, nil, nil)
}

func newWithDepPackage(name, versionStr, dep string) *api.Package {
	return newPackage(name, versionStr, []string{dep}, nil, nil, nil)
}

func pkgKey(name, versionStr string) api.PackageKey {
	return api.PackageKey{Name: name, Version: newVersion(versionStr), Arch: ""}
}

func TestLoader_Load(t *testing.T) {
	g := NewGomegaWithT(t)

	doLoad := func(
		packages []*api.Package,
		matched, ignoreRegex, allowRegex []string,
		nobest bool,
	) (*Model, *Loader) {
		loader := NewLoader()
		model, err := loader.Load(
			packages, matched, ignoreRegex, allowRegex, nobest, []string{"x86_64", "noarch"})
		g.Expect(err).ToNot(HaveOccurred())
		return model, loader
	}

	doSimpleLoad := func(packages []*api.Package, nobest bool) (*Model, *Loader) {
		return doLoad(packages, nil, nil, nil, nobest)
	}

	x1 := bf.Var("x1")
	x2 := bf.Var("x2")
	x3 := bf.Var("x3")
	x4 := bf.Var("x4")
	x5 := bf.Var("x5")

	t.Run("Trivial Loading", func(t *testing.T) {
		model, _ := doSimpleLoad([]*api.Package{}, false)

		g.Expect(model.packages).To(BeEmpty())
		g.Expect(model.vars).To(BeEmpty())
		g.Expect(model.bestPackages).To(BeEmpty())
		g.Expect(model.ands).To(BeEmpty())
		g.Expect(model.forceIgnoreWithDependencies).To(BeEmpty())
	})

	t.Run("Basic Loading and Configuration", func(t *testing.T) {
		t.Run("deduplication and default allows", func(t *testing.T) {
			pkgA := newSimplePackage("A", "1.0-1")
			model, _ := doSimpleLoad([]*api.Package{pkgA, pkgA}, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:1.0-1"},
			})
			expectedVars(g, model, "A-0:1.0-1")
			expectedBest(g, model, map[string]string{"A": "0:1.0-1"})
			expectedIgnores(g, model)
			expectedAnds(g, model,
				bf.True, // Nothing to install
			)
		})

		t.Run("only newest package with nobest=false", func(t *testing.T) {
			pkgA1 := newSimplePackage("A", "1.0-1")
			pkgA2 := newSimplePackage("A", "2.0-1")
			model, _ := doSimpleLoad([]*api.Package{pkgA1, pkgA2}, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:2.0-1"},
			})
			expectedVars(g, model, "A-0:2.0-1")
			expectedBest(g, model, map[string]string{"A": "0:2.0-1"})
			expectedIgnores(g, model)
			expectedAnds(g, model,
				bf.True, // Nothing to install
			)
		})

		t.Run("all packages with nobest=true", func(t *testing.T) {
			pkgA1 := newSimplePackage("A", "1.0-1")
			pkgA2 := newSimplePackage("A", "2.0-1")
			model, _ := doSimpleLoad([]*api.Package{pkgA1, pkgA2}, true)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:1.0-1", "0:2.0-1"},
			})
			expectedVars(g, model, "A-0:1.0-1", "A-0:2.0-1")
			expectedBest(g, model, map[string]string{"A": "0:2.0-1"})
			expectedIgnores(g, model)
			expectedAnds(g, model,
				bf.Not(bf.And(x1, x2)), // No more than one `A`
			)
		})
	})

	t.Run("Filtering and Exclusion Logic", func(t *testing.T) {
		pkgA := newWithDepPackage("pkg-a", "1.0", "dep")
		pkgB := newWithDepPackage("pkg-b", "1.0", "dep")
		pkgC := newWithDepPackage("pkg-c", "1.0", "dep")

		basicExpectations := func(m *Model) {
			expectedPackages(g, m, map[string][]string{
				"pkg-a": []string{"0:1.0"},
				"pkg-b": []string{"0:1.0"},
				"pkg-c": []string{"0:1.0"},
			})
			expectedVars(
				g,
				m,
				"pkg-a-0:1.0",
				"pkg-b-0:1.0",
				"pkg-c-0:1.0",
			)
			expectedBest(g, m, map[string]string{
				"pkg-a": "0:1.0",
				"pkg-b": "0:1.0",
				"pkg-c": "0:1.0",
			})
		}

		t.Run("filter with allowRegex only", func(t *testing.T) {
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, []string{"pkg-a"}, nil, []string{"^pkg-[ab]"}, false)

			basicExpectations(model)
			expectedIgnores(g, model, pkgKey("pkg-c", "0:1.0"))
			expectedAnds(g, model,
				bf.False, // Can't install any package (missing dependency `dep`).
			)

			// verify side effect
			g.Expect(packages[2].Format.Requires.Entries).To(BeNil())
		})

		t.Run("multiple allowRegex", func(t *testing.T) {
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, nil, []string{"^pkg-a", "^pkg-c"}, false)

			basicExpectations(model)
			expectedIgnores(g, model, pkgKey("pkg-b", "0:1.0"))
			expectedAnds(g, model,
				bf.Not(x1), // Can't install package `pkg-a` (missing dependency `dep`).
			)

			// verify side effect
			g.Expect(packages[2].Format.Requires.Entries).To(BeNil())
		})

		t.Run("filter with ignoreRegex only", func(t *testing.T) {
			pkgC.Format.Requires.Entries = []api.Entry{{Name: "dep"}}
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, []string{"^pkg-b", "^pkg-x"}, nil, false)

			basicExpectations(model)
			expectedIgnores(g, model, pkgKey("pkg-b", "0:1.0"))
			expectedAnds(g, model,
				bf.Not(x1), // Can't install package `pkg-a` (missing dependency `dep`).
				bf.Not(x3), // Can't install package `pkg-c` (missing dependency `dep`).
			)

			// verify side effect
			g.Expect(packages[1].Format.Requires.Entries).To(BeNil())
		})

		t.Run("filter with ignoreRegex only", func(t *testing.T) {
			pkgC.Format.Requires.Entries = []api.Entry{{Name: "dep"}}
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, []string{"^pkg-b", "^pkg-c"}, nil, false)

			basicExpectations(model)
			expectedIgnores(g, model, pkgKey("pkg-b", "0:1.0"), pkgKey("pkg-c", "0:1.0"))
			expectedAnds(g, model,
				bf.Not(x1), // Can't install package `pkg-a` (missing dependency `dep`).
			)

			// verify side effect
			g.Expect(packages[1].Format.Requires.Entries).To(BeNil())
		})

		t.Run("handle interaction between allow and ignore", func(t *testing.T) {
			pkgB.Format.Requires.Entries = []api.Entry{{Name: "dep"}}
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, []string{"^pkg-b"}, []string{"^pkg-[ab]"}, false)
			basicExpectations(model)
			expectedIgnores(g, model, pkgKey("pkg-b", "0:1.0"), pkgKey("pkg-c", "0:1.0"))
			expectedAnds(g, model,
				bf.Not(x1), // Can't install package `pkg-a` (missing dependency `dep`).
			)
		})
	})

	t.Run("Advanced Dependency Resolution", func(t *testing.T) {
		t.Run("handle file-based dependencies", func(t *testing.T) {
			pkgApp := newWithDepPackage("app", "1.0", "/usr/bin/tool")
			pkgTool := newPackage("toolkit", "2.0", nil, nil, nil, []string{"/usr/bin/tool"})

			model, _ := doLoad([]*api.Package{pkgApp, pkgTool}, []string{"app"}, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"app":     []string{"0:1.0"},
				"toolkit": []string{"0:2.0"},
			})
			expectedVars(
				g,
				model,
				"app-0:1.0",     // x1
				"toolkit-0:2.0", // x2
			)
			expectedBest(g, model, map[string]string{
				"app":     "0:1.0",
				"toolkit": "0:2.0",
			})
			expectedIgnores(g, model)
			expectedAnds(g, model,
				x1,                 // Install: app
				bf.Implies(x1, x2), // Requirement: app => toolkit
			)
		})

		t.Run("handle ambiguous providers", func(t *testing.T) {
			pkgApp := newWithDepPackage("app", "1.0", "webserver")
			pkgApache := newPackage("apache", "2.4", nil, []string{"webserver"}, nil, nil)
			pkgNginx := newPackage("nginx", "1.2", nil, []string{"webserver"}, nil, nil)
			model, _ := doLoad([]*api.Package{pkgApp, pkgApache, pkgNginx}, []string{"app"}, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"app":    []string{"0:1.0"},
				"apache": []string{"0:2.4"},
				"nginx":  []string{"0:1.2"},
			})
			// app, apache, nginx, + 2 virtual 'webserver'
			expectedVars(
				g,
				model,
				"apache-0:2.4", // x1
				"app-0:1.0",    // x2
				"nginx-0:1.2",  // x3
			)
			expectedBest(g, model, map[string]string{
				"app":    "0:1.0",
				"apache": "0:2.4",
				"nginx":  "0:1.2",
			})
			expectedIgnores(g, model)
			expectedAnds(g, model,
				x2,                            // Install: app
				bf.Implies(x2, bf.Or(x1, x3)), // Requirement: app => apache or nginx
			)
		})

		t.Run("should handle transitive dependencies", func(t *testing.T) {
			pkgA := newWithDepPackage("A", "1.0", "B")
			pkgB := newWithDepPackage("B", "1.0", "C")
			pkgC := newSimplePackage("C", "1.0")

			model, _ := doLoad([]*api.Package{pkgA, pkgB, pkgC}, nil, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:1.0"},
				"B": []string{"0:1.0"},
				"C": []string{"0:1.0"},
			})
			expectedVars(g, model, "A-0:1.0", "B-0:1.0", "C-0:1.0")
			expectedBest(g, model, map[string]string{
				"A": "0:1.0",
				"B": "0:1.0",
				"C": "0:1.0",
			})
			expectedAnds(g, model,
				bf.Implies(x1, x2), // Requirement: a => b
				bf.Implies(x2, x3), // Requirement: b => c
			)
		})
	})

	t.Run("Edge Cases and Robustness", func(t *testing.T) {
		t.Run("should correctly handle reducer.FixPackages", func(t *testing.T) {
			pkg := newWithDepPackage("platform-python", "3.6", "/usr/libexec/platform-python")
			model, _ := doLoad([]*api.Package{pkg}, nil, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"platform-python": []string{"0:3.6"},
			})
			expectedVars(
				g,
				model,
				"platform-python-0:3.6",
			)
			expectedBest(
				g,
				model,
				map[string]string{"platform-python": "0:3.6"},
			)
			expectedIgnores(g, model)
			expectedAnds(g, model,
				bf.True, // Nothing to install
			)

			// verify side effect
			g.Expect(pkg.Format.Requires.Entries).To(BeEmpty())
			g.Expect(pkg.Format.Provides.Entries).To(
				ContainElement(api.Entry{Name: "/usr/libexec/platform-python"}))
		})

		t.Run("should correctly handle complex version comparisons", func(t *testing.T) {
			pkgV1 := newSimplePackage("A", "1:2.0-1")
			pkgV2 := newSimplePackage("A", "5:1.0-1~rc1")
			pkgV3 := newSimplePackage("A", "5:1.0-2")
			model, _ := doLoad([]*api.Package{pkgV1, pkgV2, pkgV3}, nil, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"5:1.0-2"},
			})
			expectedVars(g, model, "A-5:1.0-2")
			expectedBest(g, model, map[string]string{"A": "5:1.0-2"})
			expectedAnds(g, model,
				bf.True, // Nothing to install
			)
		})

		t.Run("should handle empty rel provides", func(t *testing.T) {
			pkg0 := newPackage("pkgX", "1.0", []string{"gcc", "gcc EQ 11.0-xyz"}, nil, nil, nil)
			pkg1 := newPackage("gcc", "11.0", []string{"gcc11 EQ 11.0"}, []string{"gcc EQ 11.0-xyz"}, nil, nil)
			pkg2 := newPackage("gcc11", "11.0", nil, []string{"gcc EQ 11.0", "gcc11 EQ 11.0"}, nil, nil)
			allPkgs := []*api.Package{pkg0, pkg1, pkg2}
			model, _ := doLoad(allPkgs, []string{"pkgX"}, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"pkgX":  []string{"0:1.0"},
				"gcc":   []string{"0:11.0"},
				"gcc11": []string{"0:11.0"},
			})
			expectedVars(
				g,
				model,
				"gcc-0:11.0",   // x1
				"gcc11-0:11.0", // x2
				"pkgX-0:1.0",   // x3
			)
			expectedBest(g, model, map[string]string{
				"gcc":   "0:11.0",
				"gcc11": "0:11.0",
				"pkgX":  "0:1.0",
			})

			expectedAnds(g, model,
				x3,                            // Install: pkgX
				bf.Implies(x3, bf.Or(x1, x2)), // Requirement: pkgX => gcc or gcc11
				bf.Implies(x1, x2),            // Requirement: gcc => gcc11
			)

		})

		t.Run("should ignore self-conflicts", func(t *testing.T) {
			pkgA := newPackage("A", "1.0", nil, nil, []string{"A"}, nil)
			model, _ := doLoad([]*api.Package{pkgA}, nil, nil, nil, false)

			expectedVars(g, model, "A-0:1.0")
			expectedAnds(g, model,
				bf.True, // Nothing to install
			)
		})
	})

	t.Run("Error Handling Scenarios", func(t *testing.T) {
		t.Run("should handle unsatisfiable requirements", func(t *testing.T) {
			pkgA := newWithDepPackage("A", "1.0-1", "B")
			model, _ := doLoad([]*api.Package{pkgA}, nil, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:1.0-1"},
			})
			expectedVars(g, model, "A-0:1.0-1")
			expectedAnds(g, model,
				bf.Not(x1), // Can't install package `A` (missing dependency `B`).
			)
		})

		t.Run("should handle missing matched packages", func(t *testing.T) {
			pkgA := newSimplePackage("A", "1.0")
			loader := NewLoader()
			model, err := loader.Load([]*api.Package{pkgA}, []string{"non-existent"}, nil, nil, false, []string{"x86_64", "noarch"})

			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("package non-existent does not exist"))
			g.Expect(model).To(BeNil())
		})
	})

	t.Run("Repo priority comparison", func(t *testing.T) {
		testCases := []struct {
			constraint   string
			pkgAVersion  string
			pkgAPriority int
			pkgBVersion  string
			pkgBPriority int
			selected     string
		}{
			{"same version, same repo priority", "1.0", 1, "1.0", 1, "A"},
			{"same version, lower repo priority", "1.0", 2, "1.0", 1, "B"},
			{"same version, higher repo priority", "1.0", 1, "1.0", 2, "A"},
			{"higher version, same repo priority", "2.0", 1, "1.0", 1, "A"},
			{"higher version, lower repo priority", "2.0", 2, "1.0", 1, "B"},
			{"higher version, lower repo priority 2", "2.0", 5, "1.0", 1, "B"},
			{"higher version, higher repo priority", "2.0", 1, "1.0", 2, "A"},
			{"lower version, same repo priority", "1.0", 1, "2.0", 1, "B"},
			{"lower version, lower repo priority", "1.0", 2, "2.0", 1, "B"},
			{"lower version, higher repo priority", "1.0", 1, "2.0", 2, "A"},
			{"lower version, higher repo priority 2", "1.0", 1, "2.0", 3, "A"},
		}

		for _, tc := range testCases {
			t.Run(tc.constraint, func(t *testing.T) {
				pkgA := newSimplePackage("X", tc.pkgAVersion)
				pkgA.Repository.Priority = tc.pkgAPriority
				pkgB := newSimplePackage("X", tc.pkgBVersion)
				pkgB.Repository.Priority = tc.pkgBPriority
				model, _ := doSimpleLoad([]*api.Package{pkgA, pkgB}, false)

				both := map[string]*api.Package{
					"A": pkgA,
					"B": pkgB,
				}

				selectedVersion := both[tc.selected].Version.String()
				expectedPackages(g, model, map[string][]string{
					"X": []string{selectedVersion},
				})
				expectedVars(g, model, "X-"+selectedVersion)
				expectedBest(g, model, map[string]string{
					"X": selectedVersion,
				})
				expectedIgnores(g, model)
				expectedAnds(g, model,
					bf.True, // Nothing to install
				)
			})
		}
	})

	t.Run("Version Comparison Operators", func(t *testing.T) {
		allB := []*api.Package{}
		for _, v := range []string{
			"1.0-1", // x2
			"2.0-1", // x3
			"2.0-2", // x4
			"3.0-1", // x5
		} {
			allB = append(allB, newSimplePackage("B", v))
		}

		atMostOneB := bf.Not(bf.Or(
			bf.And(x2, x3),
			bf.And(x2, x4),
			bf.And(x2, x5),
			bf.And(x3, x4),
			bf.And(x3, x5),
			bf.And(x4, x5),
		))

		eqExpectedAnds := []bf.Formula{
			x1,                 // Install: A
			bf.Implies(x1, x4), // Requirement: A => B eq 2.0-2
			atMostOneB,
		}
		gtExpectedAnds := []bf.Formula{
			x1,                 // Install: A
			bf.Implies(x1, x5), // Requirement: A => B gt 2.0-2
			atMostOneB,
		}
		ltExpectedAnds := []bf.Formula{
			x1,                            // Install: A
			bf.Implies(x1, bf.Or(x2, x3)), // Requirement: A => B lt 2.0-2
			atMostOneB,
		}
		geExpectedAnds := []bf.Formula{
			x1,                            // Install: A
			bf.Implies(x1, bf.Or(x4, x5)), // Requirement: A => B ge 2.0-2
			atMostOneB,
		}
		leExpectedAnds := []bf.Formula{
			x1,                                // Install: A
			bf.Implies(x1, bf.Or(x2, x3, x4)), // Requirement: A => B le 2.0-2
			atMostOneB,
		}

		testCases := []struct {
			constraint   string
			expectedAnds []bf.Formula
		}{
			{"EQ", eqExpectedAnds},
			{"GT", gtExpectedAnds},
			{"LT", ltExpectedAnds},
			{"GE", geExpectedAnds},
			{"LE", leExpectedAnds},
		}

		for _, tc := range testCases {
			t.Run(tc.constraint, func(t *testing.T) {
				req := "B " + tc.constraint + " 2.0-2"
				pkgA := newWithDepPackage("A", "1.0-1", req)
				model, _ := doLoad(append(allB, pkgA), []string{"A"}, nil, nil, true)
				expectedAnds(g, model, append(tc.expectedAnds, x1)...)
			})
		}
	})
}
