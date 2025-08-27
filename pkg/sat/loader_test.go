package sat

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/crillab/gophersat/bf"
	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/zyedidia/generic/set"
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

func expectedIgnores(g *WithT, m *Model, pkgNames... string) {
	g.Expect(m.forceIgnoreWithDependencies).To(HaveLen(len(pkgNames)))

	for _, name := range pkgNames {
		g.Expect(m.forceIgnoreWithDependencies[name].String()).To(Equal(name))
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
		Name:    name,
		Version: newVersion(versionStr),
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

func findVarByName(m *Model, name, version string) *Var {
	for _, v := range m.vars {
		if v.Package.Name == name &&
			v.Package.Version.String() == version &&
			v.varType == VarTypePackage {
			return v
		}
	}
	return nil
}

func findFileVarByName(m *Model, path string) *Var {
	for _, v := range m.vars {
		if v.Context.Provides == path && v.varType == VarTypeFile {
			return v
		}
	}
	return nil
}

// hasUniqueConstraint checks if a model's formula contains a Unique constraint.
func hasUniqueConstraint(m *Model, varNames ...string) bool {
	// This regex finds bf.Unique clauses
	re := regexp.MustCompile(`and\(and\(or\(.*\), or\(.*\)\)\)`)

	// This regex matches the variables in bf.Unique clauses
	varRE := regexp.MustCompile(`x[0-9]+(, x[0-9]+)+`)

	// Create a set of the variable names we expect to find
	expected := set.NewMapset[string](varNames...)

	for _, formula := range m.ands {
		for _, match := range re.FindAllStringSubmatch(formula.String(), -1) {
			varMatches := varRE.FindAllStringSubmatch(match[0], -1)
			actual := set.NewMapset[string](strings.Split(varMatches[0][0], ", ")...)

			// Check if all of the actual and expected variables are the same
			if actual.Size() == expected.Size() &&
				expected.Difference(actual).Size() == 0 {
				return true
			}
		}
	}

	return false
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
			packages, matched, ignoreRegex, allowRegex, nobest)
		g.Expect(err).ToNot(HaveOccurred())
		return model, loader
	}

	doSimpleLoad := func(packages []*api.Package, nobest bool) (*Model, *Loader) {
		return doLoad(packages, nil, nil, nil, nobest)
	}

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
			expectedVars(g, model, "A-0:1.0-1(A)")
			expectedBest(g, model, map[string]string{"A": "0:1.0-1"})
			expectedIgnores(g, model)
		})

		t.Run("only newest package with nobest=false", func(t *testing.T) {
			pkgA1 := newSimplePackage("A", "1.0-1")
			pkgA2 := newSimplePackage("A", "2.0-1")
			model, _ := doSimpleLoad([]*api.Package{pkgA1, pkgA2}, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:2.0-1"},
			})
			expectedVars(g, model, "A-0:2.0-1(A)")
			expectedBest(g, model, map[string]string{"A": "0:2.0-1"})
			expectedIgnores(g, model)
		})

		t.Run("all packages with nobest=true", func(t *testing.T) {
			pkgA1 := newSimplePackage("A", "1.0-1")
			pkgA2 := newSimplePackage("A", "2.0-1")
			model, _ := doSimpleLoad([]*api.Package{pkgA1, pkgA2}, true)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:1.0-1", "0:2.0-1"},
			})
			expectedVars(g, model, "A-0:1.0-1(A)", "A-0:2.0-1(A)")
			expectedBest(g, model, map[string]string{"A": "0:2.0-1"})
			expectedIgnores(g, model)
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
				"pkg-a-0:1.0(pkg-a)",
				"pkg-b-0:1.0(pkg-b)",
				"pkg-c-0:1.0(pkg-c)",
			)
			expectedBest(g, m, map[string]string{
				"pkg-a": "0:1.0",
				"pkg-b": "0:1.0",
				"pkg-c": "0:1.0",
			})
		}

		t.Run("filter with allowRegex only", func(t *testing.T) {
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, nil, []string{"^pkg-[ab]"}, false)

			basicExpectations(model)
			expectedIgnores(g, model, "pkg-c-0:1.0")

			// verify side effect
			g.Expect(packages[2].Format.Requires.Entries).To(BeNil())
		})

		t.Run("multiple allowRegex", func(t *testing.T) {
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, nil, []string{"^pkg-a", "^pkg-c"}, false)

			basicExpectations(model)
			expectedIgnores(g, model, "pkg-b-0:1.0")

			// verify side effect
			g.Expect(packages[2].Format.Requires.Entries).To(BeNil())
		})

		t.Run("filter with ignoreRegex only", func(t *testing.T) {
			pkgC.Format.Requires.Entries = []api.Entry{{Name: "dep"}}
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, []string{"^pkg-b", "^pkg-x"}, nil, false)

			basicExpectations(model)
			expectedIgnores(g, model, "pkg-b-0:1.0")

			// verify side effect
			g.Expect(packages[1].Format.Requires.Entries).To(BeNil())
		})

		t.Run("filter with ignoreRegex only", func(t *testing.T) {
			pkgC.Format.Requires.Entries = []api.Entry{{Name: "dep"}}
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, []string{"^pkg-b", "^pkg-c"}, nil, false)

			basicExpectations(model)
			expectedIgnores(g, model, "pkg-b-0:1.0", "pkg-c-0:1.0")

			// verify side effect
			g.Expect(packages[1].Format.Requires.Entries).To(BeNil())
		})

		t.Run("handle interaction between allow and ignore", func(t *testing.T) {
			pkgB.Format.Requires.Entries = []api.Entry{{Name: "dep"}}
			packages := []*api.Package{pkgA, pkgB, pkgC}

			model, _ := doLoad(packages, nil, []string{"^pkg-b"}, []string{"^pkg-[ab]"}, false)
			basicExpectations(model)
			expectedIgnores(g, model, "pkg-b-0:1.0", "pkg-c-0:1.0")
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
				"app-0:1.0(app)",
				"toolkit-0:2.0(toolkit)",
				"toolkit-0:2.0(/usr/bin/tool)",
			)
			expectedBest(g, model, map[string]string{
				"app": "0:1.0",
				"toolkit": "0:2.0",
			})
			expectedIgnores(g, model)

			varApp := findVarByName(model, "app", "0:1.0")
			varFile := findFileVarByName(model, "/usr/bin/tool")
			g.Expect(varFile).ToNot(BeNil())
			expectedRule := bf.Implies(bf.Var(varApp.satVarName), bf.And(bf.Unique(varFile.satVarName), bf.Var(varApp.satVarName)))
			g.Expect(model.ands).To(ContainElement(expectedRule))
		})

		t.Run("handle ambiguous providers", func(t *testing.T) {
			pkgApp := newWithDepPackage("app", "1.0", "webserver")
			pkgApache := newPackage("apache", "2.4", nil, []string{"webserver"}, nil, nil)
			pkgNginx := newPackage("nginx", "1.2", nil, []string{"webserver"}, nil, nil)
			model, loader := doLoad([]*api.Package{pkgApp, pkgApache, pkgNginx}, []string{"app"}, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"app":    []string{"0:1.0"},
				"apache": []string{"0:2.4"},
				"nginx":  []string{"0:1.2"},
			})
			// app, apache, nginx, + 2 virtual 'webserver'
			expectedVars(
				g,
				model,
				"apache-0:2.4(apache)",
				"apache-0:2.4(webserver)",
				"app-0:1.0(app)",
				"nginx-0:1.2(nginx)",
				"nginx-0:1.2(webserver)",
			)
			expectedBest(g, model, map[string]string{
				"app": "0:1.0",
				"apache": "0:2.4",
				"nginx": "0:1.2",
			})
			expectedIgnores(g, model)
			apacheVar := loader.provides["webserver"][0]
			nginxVar := loader.provides["webserver"][1]
			g.Expect(hasUniqueConstraint(model, apacheVar.satVarName, nginxVar.satVarName)).To(BeTrue())
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
			expectedVars(g, model, "A-0:1.0(A)", "B-0:1.0(B)", "C-0:1.0(C)")
			expectedBest(g, model, map[string]string{
				"A": "0:1.0",
				"B": "0:1.0",
				"C": "0:1.0",
			})
			varA := findVarByName(model, "A", "0:1.0")
			varB := findVarByName(model, "B", "0:1.0")
			varC := findVarByName(model, "C", "0:1.0")
			g.Expect(model.ands).To(ContainElement(bf.Implies(
				bf.Var(varA.satVarName),
				bf.And(bf.Unique(varB.satVarName), bf.Var(varA.satVarName)),
			)))
			g.Expect(model.ands).To(ContainElement(bf.Implies(
				bf.Var(varB.satVarName),
				bf.And(bf.Unique(varC.satVarName), bf.Var(varB.satVarName)),
			)))
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
				"platform-python-0:3.6(platform-python)",
				"platform-python-0:3.6(/usr/libexec/platform-python)",
			)
			expectedBest(
				g,
				model,
				map[string]string{"platform-python": "0:3.6"},
			)
			expectedIgnores(g, model)

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
			expectedVars(g, model, "A-5:1.0-2(A)")
			expectedBest(g, model, map[string]string{"A": "5:1.0-2"})
		})

		t.Run("should ignore self-conflicts", func(t *testing.T) {
			pkgA := newPackage("A", "1.0", nil, nil, []string{"A"}, nil)
			model, _ := doLoad([]*api.Package{pkgA}, nil, nil, nil, false)

			expectedVars(g, model, "A-0:1.0(A)")
			varA := findVarByName(model, "A", "0:1.0")

			// The goal is to ensure NO conflict rule is generated for a package against itself.
			// A conflict rule has the string structure: "Implies(var, Not(Or(...)))"
			foundSelfConflict := false
			for _, formula := range model.ands {
				// Convert formula to its public string representation
				formulaStr := formula.String()
				// Check if a formula starts with an implication from our variable and contains a negation.
				if strings.HasPrefix(formulaStr, "and("+varA.satVarName) && strings.Contains(formulaStr, "not(or") {
					foundSelfConflict = true
					break
				}
			}
			g.Expect(foundSelfConflict).To(BeFalse(), "A conflict rule should not be generated for a self-conflicting package")
		})
	})

	t.Run("Error Handling Scenarios", func(t *testing.T) {
		t.Run("should handle unsatisfiable requirements", func(t *testing.T) {
			pkgA := newWithDepPackage("A", "1.0-1", "B")
			model, _ := doLoad([]*api.Package{pkgA}, nil, nil, nil, false)

			expectedPackages(g, model, map[string][]string{
				"A": []string{"0:1.0-1"},
			})
			expectedVars(g, model, "A-0:1.0-1(A)")
			varA := findVarByName(model, "A", "0:1.0-1")
			expectedRule := bf.Implies(bf.Var(varA.satVarName), bf.Not(bf.Var(varA.satVarName)))
			g.Expect(model.ands).To(ContainElement(expectedRule))
		})

		t.Run("should handle missing matched packages", func(t *testing.T) {
			pkgA := newSimplePackage("A", "1.0")
			loader := NewLoader()
			model, err := loader.Load([]*api.Package{pkgA}, []string{"non-existent"}, nil, nil, false)

			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("package non-existent does not exist"))
			g.Expect(model).To(BeNil())
		})
	})

	t.Run("Version Comparison Operators", func(t *testing.T) {
		allB := []*api.Package{}
		for _, v := range []string{"1.0-1", "2.0-1", "2.0-2", "3.0-1"} {
			allB = append(allB, newSimplePackage("B", v))
		}

		testCases := []struct {
			constraint string
			satisfiers []*api.Package
			rejects    []*api.Package
		}{
			{"EQ", []*api.Package{allB[2]}, []*api.Package{allB[0], allB[1], allB[3]}},
			{"GT", []*api.Package{allB[3]}, []*api.Package{allB[0], allB[1], allB[2]}},
			{"LT", []*api.Package{allB[0], allB[1]}, []*api.Package{allB[2], allB[3]}},
			{"GE", []*api.Package{allB[2], allB[3]}, []*api.Package{allB[0], allB[1]}},
			{"LE", []*api.Package{allB[0], allB[1], allB[2]}, []*api.Package{allB[3]}},
		}

		for _, tc := range testCases {
			t.Run(tc.constraint, func(t *testing.T) {
				req := "B " + tc.constraint + " 2.0-2"
				pkgA := newWithDepPackage("A", "1.0-1", req)
				model, _ := doLoad(append(allB, pkgA), []string{"A"}, nil, nil, true)

				// --- Verification Logic ---
				// 1. Find the specific dependency formula for package A
				//    A dependency rule "A -> B" is represented as "or(not(A), B)".
				varA := findVarByName(model, "A", "0:1.0-1")
				g.Expect(varA).ToNot(BeNil())
				var dependencyFormulaStr string

				for _, formula := range model.ands {
					if strings.HasPrefix(formula.String(), "or(not("+varA.satVarName+")") &&
						strings.Contains(formula.String(), "and(and(") {
						dependencyFormulaStr = formula.String()
						break
					}
				}

				g.Expect(dependencyFormulaStr).ToNot(BeEmpty(), "Could not find the dependency formula for package A")
				// 2. Verify that expected packages are present
				for _, pkg := range tc.satisfiers {
					v := findVarByName(model, pkg.Name, pkg.Version.String())
					g.Expect(v).ToNot(BeNil())
					g.Expect(dependencyFormulaStr).To(ContainSubstring(v.satVarName))
				}

				// 3. Verify that packages were rejected
				for _, pkg := range tc.rejects {
					v := findVarByName(model, pkg.Name, pkg.Version.String())
					g.Expect(v).ToNot(BeNil())
					g.Expect(dependencyFormulaStr).ToNot(ContainSubstring(v.satVarName))
				}
			})
		}
	})
}
