package sat

import (
	"cmp"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/crillab/gophersat/bf"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Loader struct {
	m         *Model
	provides  map[string][]*Var
	varsCount int
}

// BestKey groups packages for the purpose of `--nobest` option disabled,
// so that for each set of packages sharing this key, only the "best" package will be preserved
// (best in terms of repo priority, version criteria).
type BestKey struct {
	name string
	arch string
}

// CompareBestKey provides an arbitrary, deterministic, total order on BestKey
func CompareBestKey(k1 BestKey, k2 BestKey) int {
	return cmp.Or(
		cmp.Compare(k1.name, k2.name),
		cmp.Compare(k1.arch, k2.arch),
	)
}

func MakeBestKey(pkg *api.Package) BestKey {
	return BestKey{name: pkg.Name, arch: pkg.Arch}
}

func NewLoader() *Loader {
	return &Loader{
		m: &Model{
			packages:                    map[string][]*Var{},
			vars:                        map[string]*Var{},
			bestPackages:                map[BestKey]*api.Package{},
			forceIgnoreWithDependencies: map[api.PackageKey]*api.Package{},
		},
		provides:  map[string][]*Var{},
		varsCount: 0,
	}
}

// Load takes a list of all involved packages to install, a list of regular
// expressions which denote packages which should be taken into account for
// solving the problem, but they should then be ignored together with their
// requirements in the provided list of installed packages, and also a list
// of regular expressions that may be used to limit the selection to matching
// packages.
func (loader *Loader) Load(packages []*api.Package, matched, ignoreRegex, allowRegex []string, nobest bool, archOrder []string) (*Model, error) {
	// Deduplicate and detect excludes
	deduplicated := map[api.PackageKey]*api.Package{}
	for i, pkg := range packages {
		if _, exists := deduplicated[pkg.Key()]; exists {
			logrus.Infof("Removing duplicate of  %v.", pkg.String())
		}
		fullName := pkg.MatchableString()
		if _, exists := deduplicated[pkg.Key()]; !exists {
			allowed := len(allowRegex) == 0
			for _, rex := range allowRegex {
				if match, err := regexp.MatchString(rex, fullName); err != nil {
					return nil, fmt.Errorf("failed to match package with regex '%v': %v", rex, err)
				} else if match {
					allowed = true
					break
				}
			}

			ignored := false
			if allowed {
				for _, rex := range ignoreRegex {
					if match, err := regexp.MatchString(rex, fullName); err != nil {
						return nil, fmt.Errorf("failed to match package with regex '%v': %v", rex, err)
					} else if match {
						logrus.Warnf("Package %v is forcefully ignored by regex '%v'.", pkg.String(), rex)
						ignored = true
						break
					}
				}
			}

			if !allowed {
				logrus.Warnf("Package %v is not explicitly allowed", pkg.String())
			}

			if !allowed || ignored {
				packages[i].Format.Requires.Entries = nil
				loader.m.forceIgnoreWithDependencies[pkg.Key()] = packages[i]
			}

			deduplicated[pkg.Key()] = packages[i]
		}
	}

	deduplicatedKeys := maps.Keys(deduplicated)
	slices.SortFunc(deduplicatedKeys, rpm.ComparePackageKey)

	packages = nil
	for _, k := range deduplicatedKeys {
		reducer.FixPackages(deduplicated[k])
		packages = append(packages, deduplicated[k])
	}

	// Create an index to pick the best candidates
	for _, pkg := range packages {
		key := MakeBestKey(pkg)
		if loader.m.bestPackages[key] == nil {
			loader.m.bestPackages[key] = pkg
		} else if rpm.ComparePackage(pkg, loader.m.bestPackages[key], archOrder) > 0 {
			loader.m.bestPackages[key] = pkg
		}
	}

	if !nobest {
		packages = nil
		bestPackagesKeys := maps.Keys(loader.m.bestPackages)
		slices.SortFunc(bestPackagesKeys, CompareBestKey)
		for _, v := range bestPackagesKeys {
			packages = append(packages, loader.m.bestPackages[v])
		}
	}

	pkgProvides := map[VarContext][]*Var{}

	// Generate variables
	for _, pkg := range packages {
		pkgVar, resourceVars := loader.explodePackageToVars(pkg)
		loader.m.packages[pkg.Name] = append(loader.m.packages[pkg.Name], pkgVar)
		pkgProvides[pkgVar.Context] = resourceVars
		for _, v := range resourceVars {
			loader.provides[v.Context.Provides] = append(loader.provides[v.Context.Provides], v)
			loader.m.vars[v.satVarName] = v
		}
	}

	packagesKeys := maps.Keys(loader.m.packages)
	slices.Sort(packagesKeys)

	for _, x := range packagesKeys {
		sort.SliceStable(loader.m.packages[x], func(i, j int) bool {
			return rpm.ComparePackage(loader.m.packages[x][i].Package, loader.m.packages[x][j].Package, archOrder) < 0
		})
	}

	logrus.Infof("Loaded %v packages.", len(pkgProvides))

	pkgProvideKeys := maps.Keys(pkgProvides)
	slices.SortFunc(pkgProvideKeys, varContextSort)

	// Generate imply rules
	for _, provided := range pkgProvideKeys {
		// Create imply rules for every package and add them to the formula
		// one provided dependency implies all dependencies from that package
		resourceVars := pkgProvides[provided]
		bfVar := bf.And(toBFVars(resourceVars)...)
		var ands []bf.Formula
		for _, res := range resourceVars {
			ands = append(ands, bf.Implies(bf.Var(res.satVarName), bfVar))
		}
		pkgVar := resourceVars[len(resourceVars)-1]
		ands = append(ands, bf.Implies(bf.Var(pkgVar.satVarName), loader.explodePackageRequires(pkgVar)))
		if conflicts := loader.explodePackageConflicts(pkgVar); conflicts != nil {
			ands = append(ands, bf.Implies(bf.Var(pkgVar.satVarName), bf.Not(conflicts)))
		}

		// Implicit conflicts (with the same package):
		ands = append(ands, bf.Implies(bf.Var(pkgVar.satVarName), bf.Not(loader.explodeSamePackageConflicts(pkgVar))))

		loader.m.ands = append(loader.m.ands, ands...)
	}
	logrus.Infof("Generated %v variables.", len(loader.m.vars))

	return loader.constructRequirements(matched, archOrder)
}

func (loader *Loader) explodePackageToVars(pkg *api.Package) (pkgVar *Var, resourceVars []*Var) {
	for _, p := range pkg.Format.Provides.Entries {
		if p.Name == pkg.Name {
			pkgVar = &Var{
				satVarName: loader.ticket(),
				varType:    VarTypePackage,
				Context: VarContext{
					PackageKey: pkg.Key(),
					Provides:   pkg.Name,
				},
				Package:         pkg,
				ResourceVersion: &pkg.Version,
			}
			resourceVars = append(resourceVars, pkgVar)
		} else {
			resVar := &Var{
				satVarName: loader.ticket(),
				varType:    VarTypeResource,
				Context: VarContext{
					PackageKey: pkg.Key(),
					Provides:   p.Name,
				},
				ResourceVersion: &api.Version{
					Rel:   p.Rel,
					Ver:   p.Ver,
					Epoch: p.Epoch,
				},
				Package: pkg,
			}
			resourceVars = append(resourceVars, resVar)
		}
	}

	for _, f := range pkg.Format.Files {
		resVar := &Var{
			satVarName: loader.ticket(),
			varType:    VarTypeFile,
			Context: VarContext{
				PackageKey: pkg.Key(),
				Provides:   f.Text,
			},
			Package:         pkg,
			ResourceVersion: &api.Version{},
		}
		resourceVars = append(resourceVars, resVar)
	}
	return pkgVar, resourceVars
}

// explodePackageRequires builds a formula that could be a right hand side operand to implication.
// It consists of all direct requirements of a given package, exploded to resources that can satisfy these requirements.
// Special cases include:
// - no requirements: returning `bf.True`
// - can't satisfy requirements (because of lack of providers): returning `bf.False`
// Both are safe to use in an implication.
func (loader *Loader) explodePackageRequires(pkgVar *Var) bf.Formula {
	var requirements [][]*Var
	ok := true
	for _, req := range pkgVar.Package.Format.Requires.Entries {
		satisfies, err := loader.explodeSingleRequires(req, loader.provides[req.Name])
		if err != nil {
			logrus.Warnf("Package %s requires %s, but only got %+v", pkgVar.Package, req, loader.provides[req.Name])
			ok = false
			continue
		}
		requirements = append(requirements, satisfies)
	}

	if !ok {
		return bf.False
	}

	var orRequirements []bf.Formula
	for _, satisfies := range requirements {
		vars := []bf.Formula{}
		for _, s := range satisfies {
			vars = append(vars, bf.Var(s.satVarName))
		}
		orRequirements = append(orRequirements, bf.Or(vars...))
	}

	if orRequirements == nil {
		// empty `bf.And` doesn't work as expected, hence this special case:
		return bf.True
	}
	return bf.And(orRequirements...)
}

func (loader *Loader) explodePackageConflicts(pkgVar *Var) bf.Formula {
	conflictingVars := []bf.Formula{}
	for _, req := range pkgVar.Package.Format.Conflicts.Entries {
		conflicts, err := loader.explodeSingleRequires(req, loader.provides[req.Name])
		if err != nil {
			// if a conflicting resource does not exist, we don't care
			continue
		}
		for _, s := range conflicts {
			if s.Package == pkgVar.Package {
				// don't conflict with yourself
				//logrus.Infof("%s does not conflict with %s", s.Package.String(), pkgVar.Package.String())
				continue
			}
			if !strings.HasPrefix(s.Package.Name, "fedora-release") && !strings.HasPrefix(pkgVar.Package.Name, "fedora-release") {
				logrus.Infof("%s conflicts with %s", s.Package.String(), pkgVar.Package.String())
			}
			conflictingVars = append(conflictingVars, bf.Var(s.satVarName))
		}
	}
	if len(conflictingVars) == 0 {
		return nil
	}
	return bf.Or(conflictingVars...)
}

// explodeSamePackageConflicts returns a formula indicating whether there was installed
// a package of the same name, conflicting with one represented by `pkgVar`.
func (loader *Loader) explodeSamePackageConflicts(pkgVar *Var) bf.Formula {
	var conflictingVars []bf.Formula
	for _, otherVar := range loader.m.packages[pkgVar.Package.Name] {
		if otherVar.Package == pkgVar.Package { // itself
			continue
		}
		if otherVar.Package.Arch != pkgVar.Package.Arch {
			continue
		}
		conflictingVars = append(conflictingVars, bf.Var(otherVar.satVarName))
	}
	if len(conflictingVars) == 0 {
		return bf.False
	}
	return bf.Or(conflictingVars...)
}

func (loader *Loader) explodeSingleRequires(entry api.Entry, provides []*Var) (accepts []*Var, err error) {
	provPerPkg := map[VarContext][]*Var{}
	for _, prov := range provides {
		provPerPkg[prov.Context] = append(provPerPkg[prov.Context], prov)
	}

	for _, pkgProv := range provPerPkg {
		acceptsFromPkg, err := compareRequires(entry, pkgProv)
		if err != nil {
			return nil, err
		}
		if len(acceptsFromPkg) > 0 {
			// just pick one to avoid excluding each  other
			accepts = append(accepts, acceptsFromPkg[0])
		}
	}

	if len(accepts) == 0 {
		return nil, fmt.Errorf("Nothing can satisfy %s", entry.Name)
	}

	return accepts, nil
}

func (loader *Loader) ticket() string {
	loader.varsCount++
	return "x" + strconv.Itoa(loader.varsCount)
}

func (loader *Loader) constructRequirements(packages []string, archOrder []string) (*Model, error) {
	logrus.Info("Adding required packages to the resolver.")

	for _, pkgName := range packages {
		req, err := loader.resolveNewest(pkgName, archOrder)
		if err != nil {
			return nil, err
		}
		logrus.Infof("Selecting %s: %v", pkgName, req.Package)
		loader.m.ands = append(loader.m.ands, bf.Var(req.satVarName))
	}
	return loader.m, nil
}

func (loader *Loader) resolveNewest(pkgName string, archOrder []string) (*Var, error) {
	pkgs := loader.provides[pkgName]
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s does not exist", pkgName)
	}
	newest := pkgs[0]
	for _, p := range pkgs {
		if rpm.ComparePackage(p.Package, newest.Package, archOrder) > 0 {
			newest = p
		}
	}
	return newest, nil
}

func toBFVars(vars []*Var) (bfvars []bf.Formula) {
	for _, v := range vars {
		bfvars = append(bfvars, bf.Var(v.satVarName))
	}
	return
}

func compareRequires(entry api.Entry, provides []*Var) (accepts []*Var, err error) {
	for _, dep := range provides {
		entryVer := api.Version{
			Text:  entry.Text,
			Epoch: entry.Epoch,
			Ver:   entry.Ver,
			Rel:   entry.Rel,
		}

		// Requirement "EQ 2.14" matches 2.14-5.fc33
		depVer := *dep.ResourceVersion
		if entryVer.Rel == "" {
			depVer.Rel = ""
		}

		// Provide "EQ 2.14" matches "2.14-5.fc33"
		if depVer.Rel == "" {
			entryVer.Rel = ""
		}

		works := false
		if depVer.Epoch == "" && depVer.Ver == "" && depVer.Rel == "" {
			works = true
		} else {
			cmp := rpm.Compare(depVer, entryVer)
			switch entry.Flags {
			case "EQ":
				if cmp == 0 {
					works = true
				}
				cmp = 0
			case "LE":
				if cmp <= 0 {
					works = true
				}
			case "GE":
				if cmp >= 0 {
					works = true
				}
			case "LT":
				if cmp == -1 {
					works = true
				}
			case "GT":
				if cmp == 1 {
					works = true
				}
			case "":
				return provides, nil
			default:
				return nil, fmt.Errorf("can't interprate flags value %s", entry.Flags)
			}
		}
		if works {
			accepts = append(accepts, dep)
		}
	}
	return accepts, nil
}
