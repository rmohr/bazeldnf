package sat

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/crillab/gophersat/bf"
	"github.com/crillab/gophersat/maxsat"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/reducer"
	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type VarType string

const (
	VarTypePackage  = "Package"
	VarTypeResource = "Resource"
	VarTypeFile     = "File"
)

// VarContext contains all information to create a unique identifyable hash key which can be traced back to a package
// for every resource in a yum repo
type VarContext struct {
	Package  string
	Provides string
	Version  api.Version
}

func varContextSort(a VarContext, b VarContext) int {
	return cmp.Or(
		cmp.Compare(a.Package, b.Package),
		rpm.Compare(a.Version, b.Version),
	)
}

type Var struct {
	satVarName      string
	varType         VarType
	Context         VarContext
	Package         *api.Package
	ResourceVersion *api.Version
}

func (v Var) String() string {
	return fmt.Sprintf("%s(%s)", v.Package.String(), v.Context.Provides)
}

func toBFVars(vars []*Var) (bfvars []bf.Formula) {
	for _, v := range vars {
		bfvars = append(bfvars, bf.Var(v.satVarName))
	}
	return
}

type Model struct {
	// packages contains a map which contains all pkg vars which can be looked up by package name
	// useful for creating soft clauses
	packages map[string][]*Var

	// vars contain as key an exact identifier for a provided resource and the actual SAT variable as value
	vars map[string]*Var

	bestPackages map[string]*api.Package

	ands                        []bf.Formula
	forceIgnoreWithDependencies map[string]*api.Package
}

func (m *Model) Packages() map[string][]*Var {
	return m.packages
}

func (m *Model) Var(v string) *Var {
	return m.vars[v]
}

func (m *Model) BestPackage(p string) *api.Package {
	return m.bestPackages[p]
}

func (m *Model) Ands() bf.Formula {
	return bf.And(m.ands...)
}

func (m *Model) ShouldIgnore(p string) bool {
	_, exists := m.forceIgnoreWithDependencies[p]
	return exists
}

type Loader struct {
	m         *Model
	provides  map[string][]*Var
	varsCount int
}

func (loader *Loader) Load(packages []*api.Package, matched, ignoreRegex, allowRegex []string, nobest bool) (*Model, error) {
	// Deduplicate and detect excludes
	deduplicated := map[string]*api.Package{}
	for i, pkg := range packages {
		if _, exists := deduplicated[pkg.String()]; exists {
			logrus.Infof("Removing duplicate of  %v.", pkg.String())
		}
		fullName := pkg.String()
		if _, exists := deduplicated[fullName]; !exists {
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
				loader.m.forceIgnoreWithDependencies[pkg.String()] = packages[i]
			}

			deduplicated[pkg.String()] = packages[i]
		}
	}

	deduplicatedKeys := maps.Keys(deduplicated)
	slices.Sort(deduplicatedKeys)

	packages = nil
	for _, k := range deduplicatedKeys {
		reducer.FixPackages(deduplicated[k])
		packages = append(packages, deduplicated[k])
	}

	// Create an index to pick the best candidates
	for _, pkg := range packages {
		if loader.m.bestPackages[pkg.Name] == nil {
			loader.m.bestPackages[pkg.Name] = pkg
		} else if rpm.Compare(pkg.Version, loader.m.bestPackages[pkg.Name].Version) == 1 {
			loader.m.bestPackages[pkg.Name] = pkg
		}
	}

	if !nobest {
		packages = nil
		bestPackagesKeys := maps.Keys(loader.m.bestPackages)
		slices.Sort(bestPackagesKeys)
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
			return rpm.Compare(loader.m.packages[x][i].Package.Version, loader.m.packages[x][j].Package.Version) < 0
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
		loader.m.ands = append(loader.m.ands, ands...)
	}
	logrus.Infof("Generated %v variables.", len(loader.m.vars))

	return loader.m, loader.constructRequirements(matched)
}

func (loader *Loader) explodePackageToVars(pkg *api.Package) (pkgVar *Var, resourceVars []*Var) {
	for _, p := range pkg.Format.Provides.Entries {
		if p.Name == pkg.Name {
			pkgVar = &Var{
				satVarName: loader.ticket(),
				varType:    VarTypePackage,
				Context: VarContext{
					Package:  pkg.Name,
					Provides: pkg.Name,
					Version:  pkg.Version,
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
					Package:  pkg.Name,
					Provides: p.Name,
					Version:  pkg.Version,
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
				Package:  pkg.Name,
				Provides: f.Text,
				Version:  pkg.Version,
			},
			Package:         pkg,
			ResourceVersion: &api.Version{},
		}
		resourceVars = append(resourceVars, resVar)
	}
	return pkgVar, resourceVars
}

func (loader *Loader) explodePackageRequires(pkgVar *Var) bf.Formula {
	var bfunique = bf.Var(pkgVar.satVarName)
	for _, req := range pkgVar.Package.Format.Requires.Entries {
		satisfies, err := loader.explodeSingleRequires(req, loader.provides[req.Name])
		if err != nil {
			logrus.Warnf("Package %s requires %s, but only got %+v", pkgVar.Package, req, loader.provides[req.Name])
			return bf.Not(bfunique)
		}
		uniqueVars := []string{}
		for _, s := range satisfies {
			uniqueVars = append(uniqueVars, s.satVarName)
		}
		bfunique = bf.And(bf.Unique(uniqueVars...), bfunique)
	}
	return bfunique
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
			if !strings.HasPrefix(s.Package.Name, "fedora-release") && !strings.HasPrefix(pkgVar.Package.String(), "fedora-release") {
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

func (loader *Loader) explodeSingleRequires(entry api.Entry, provides []*Var) (accepts []*Var, err error) {
	entryVer := api.Version{
		Text:  entry.Text,
		Epoch: entry.Epoch,
		Ver:   entry.Ver,
		Rel:   entry.Rel,
	}

	provPerPkg := map[VarContext][]*Var{}
	for _, prov := range provides {
		provPerPkg[prov.Context] = append(provPerPkg[prov.Context], prov)
	}

	for _, pkgProv := range provPerPkg {
		acceptsFromPkg, err := compareRequires(entryVer, entry.Flags, pkgProv)
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

func (loader *Loader) constructRequirements(packages []string) error {
	logrus.Info("Adding required packages to the resolver.")

	for _, pkgName := range packages {
		req, err := loader.resolveNewest(pkgName)
		if err != nil {
			return err
		}
		logrus.Infof("Selecting %s: %v", pkgName, req.Package)
		loader.m.ands = append(loader.m.ands, bf.Var(req.satVarName))
	}
	return nil
}

func (loader *Loader) resolveNewest(pkgName string) (*Var, error) {
	pkgs := loader.provides[pkgName]
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s does not exist", pkgName)
	}
	newest := pkgs[0]
	for _, p := range pkgs {
		if rpm.Compare(p.Package.Version, newest.Package.Version) == 1 {
			newest = p
		}
	}
	return newest, nil
}

type Resolver struct {
	loader *Loader
}

func NewResolver() *Resolver {
	return &Resolver{
		loader: &Loader{
			m: &Model{
				packages:                    map[string][]*Var{},
				vars:                        map[string]*Var{},
				bestPackages:                map[string]*api.Package{},
				forceIgnoreWithDependencies: map[string]*api.Package{},
			},
			provides:  map[string][]*Var{},
			varsCount: 0,
		},
	}
}

// LoadInvolvedPackages takes a list of all involved packages to install, a list of regular expressions
// which denote packages which should be taken into account for solving the problem, but they should
// then be ignored together with their requirements in the provided list of installed packages, and also
// a list of regular expressions that may be used to limit the selection to matching packages.
func (r *Resolver) LoadInvolvedPackages(packages []*api.Package, matched, ignoreRegex, allowRegex []string, nobest bool) (*Model, error) {
	return r.loader.Load(packages, matched, ignoreRegex, allowRegex, nobest)
}

func (res *Resolver) Resolve(model *Model) (install []*api.Package, excluded []*api.Package, forceIgnoredWithDependencies []*api.Package, err error) {
	logrus.WithField("bf", model.Ands()).Debug("Formula to solve")

	satReader, satWriter := io.Pipe()
	pwMaxSatReader, pwMaxSatWriter := io.Pipe()
	rex := regexp.MustCompile("c (x[0-9]+)=([0-9]+)")

	satErrChan := make(chan error, 1)
	pwMaxSatErrChan := make(chan error, 1)
	varsChan := make(chan ConversionVars, 1)
	go func() {
		defer close(satErrChan)
		defer satWriter.Close()
		satErrChan <- bf.Dimacs(model.Ands(), satWriter)
	}()

	go func() {
		defer close(pwMaxSatErrChan)
		defer pwMaxSatWriter.Close()
		vars := ConversionVars{
			satToPkg: map[string]string{},
			pkgToSat: map[string]string{},
		}
		defer func() { varsChan <- vars }()
		scanner := bufio.NewScanner(satReader)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "c") {
				match := rex.FindStringSubmatch(line)
				if len(match) == 3 {
					pkgVar := match[1]
					satVar := match[2]
					vars.satToPkg[satVar] = pkgVar
					vars.pkgToSat[pkgVar] = satVar
					if _, err := fmt.Fprintf(pwMaxSatWriter, "c %s -> %s\n", model.Var(pkgVar).Package.String(), model.Var(pkgVar).Context.Provides); err != nil {
						pwMaxSatErrChan <- err
						return
					}
				}
			} else if strings.HasPrefix(line, "p") {
				line = strings.Replace(line, "p cnf", "p wcnf", 1) + " 2000"
			} else {
				line = "2000 " + line
			}
			if _, err := fmt.Fprintln(pwMaxSatWriter, line); err != nil {
				pwMaxSatErrChan <- err
				return
			}
		}
		// write soft rules. We don't want to install any package
		for _, pkgs := range model.Packages() {
			weight := 1901
			fmt.Fprintf(pwMaxSatWriter, "c prefer %s\n", pkgs[len(pkgs)-1].Package.String())
			if len(pkgs) > 1 {
				for _, pkg := range pkgs[0 : len(pkgs)-1] {
					pkgVar := pkg.satVarName
					satVar := vars.pkgToSat[pkgVar]
					fmt.Fprintf(pwMaxSatWriter, "c not %s,%s,%s\n", pkg.Package.String(), pkgVar, satVar)
					fmt.Fprintf(pwMaxSatWriter, "%d -%s 0\n", weight, satVar)

					if weight > 0 {
						weight -= 100
					}
				}
			}
		}
	}()

	logrus.Info("Loading the Partial weighted MAXSAT problem.")
	s, err := maxsat.ParseWCNF(pwMaxSatReader)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := <-satErrChan; err != nil {
		return nil, nil, nil, err
	}
	if err := <-pwMaxSatErrChan; err != nil {
		return nil, nil, nil, err
	}
	satVars := <-varsChan

	logrus.Info("Solving the Partial weighted MAXSAT problem.")
	solution := s.Optimal(nil, nil)

	if solution.Status.String() == "SAT" {
		logrus.Infof("Solution with weight %v found.", solution.Weight)
		installMap := map[VarContext]*api.Package{}
		excludedMap := map[VarContext]*api.Package{}
		forceIgnoreMap := map[VarContext]*api.Package{}
		for k, v := range solution.Model {
			// Offset of `1`. The model index starts with 0, but the variable sequence starts with 1, since 0 is not allowed
			resVar := model.Var(satVars.satToPkg[strconv.Itoa(k+1)])
			if resVar != nil && resVar.varType == VarTypePackage {
				if v {
					if exists := model.ShouldIgnore(resVar.Package.String()); !exists {
						installMap[resVar.Context] = resVar.Package
					} else {
						forceIgnoreMap[resVar.Context] = resVar.Package
					}
				} else {
					excludedMap[resVar.Context] = resVar.Package
				}
			}
		}
		for _, v := range installMap {
			if rpm.Compare(model.BestPackage(v.Name).Version, v.Version) != 0 {
				logrus.Infof("Picking %v instead of best candiate %v", v, model.BestPackage(v.Name))
			}
			install = append(install, v)
		}

		for _, v := range excludedMap {
			excluded = append(excluded, v)
		}
		for _, v := range forceIgnoreMap {
			forceIgnoredWithDependencies = append(forceIgnoredWithDependencies, v)
		}
		return install, excluded, forceIgnoredWithDependencies, nil
	}
	logrus.Info("No solution found.")
	return nil, nil, nil, fmt.Errorf("no solution found")
}

func compareRequires(entryVer api.Version, flag string, provides []*Var) (accepts []*Var, err error) {
	for _, dep := range provides {

		// Requirement "EQ 2.14" matches 2.14-5.fc33
		depVer := *dep.ResourceVersion
		if entryVer.Rel == "" {
			depVer.Rel = ""
		}

		works := false
		if depVer.Epoch == "" && depVer.Ver == "" && depVer.Rel == "" {
			works = true
		} else {
			cmp := rpm.Compare(depVer, entryVer)
			switch flag {
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
				return nil, fmt.Errorf("can't interprate flags value %s", flag)
			}
		}
		if works {
			accepts = append(accepts, dep)
		}
	}
	return accepts, nil
}

type ConversionVars struct {
	satToPkg map[string]string
	pkgToSat map[string]string
}
