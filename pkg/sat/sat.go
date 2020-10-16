package sat

import (
	"fmt"
	"io"
	"strconv"

	"github.com/crillab/gophersat/explain"
	"github.com/rmohr/bazel-dnf/pkg/api"
	"github.com/rmohr/bazel-dnf/pkg/rpm"
	"github.com/sirupsen/logrus"

	"github.com/crillab/gophersat/bf"
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

type Var struct {
	satVarName string
	varType    VarType
	Context    VarContext
	Package    *api.Package
}

func toBFVars(vars []*Var) (bfvars []bf.Formula) {
	for _, v := range vars {
		bfvars = append(bfvars, bf.Var(v.satVarName))
	}
	return
}

type Resolver struct {
	varsCount int
	// provides allows accessing variables which can resolve unversioned requirement to build proper clauses
	provides map[string][]*Var
	// pkgProvides allows accessing all variables which get pulled in if a specific package get's pulled in
	pkgProvides map[VarContext][]*Var
	// vars contain as key an exact identifier for a provided resource and the actual SAT variable as value
	vars map[string]*Var

	bestPackages map[string]*api.Package

	ands         []bf.Formula
	unresolvable []api.Entry
	nobest       bool
}

func NewResolver(nobest bool) *Resolver {
	return &Resolver{
		varsCount:   0,
		provides:    map[string][]*Var{},
		vars:        map[string]*Var{},
		pkgProvides: map[VarContext][]*Var{},
		nobest:      nobest,
		bestPackages: map[string]*api.Package{},
	}
}

func (r *Resolver) ticket() string {
	r.varsCount++
	return strconv.Itoa(r.varsCount)
}

func (r *Resolver) LoadInvolvedPackages(packages []*api.Package) error {
	// Create an index to pick the best candidates
	for _, pkg := range packages {
		if r.bestPackages[pkg.Name] == nil {
			r.bestPackages[pkg.Name] = pkg
		} else if rpm.Compare(pkg.Version, r.bestPackages[pkg.Name].Version) == 1 {
			r.bestPackages[pkg.Name] = pkg
		}
	}

	if !r.nobest {
		packages = nil
		for _, v := range r.bestPackages {
			packages = append(packages, v)
		}
	}
	// Generate variables
	for _, pkg := range packages {
		pkgVar, resourceVars := r.explodePackageToVars(pkg)
		r.pkgProvides[pkgVar.Context] = resourceVars
		for _, v := range resourceVars {
			r.provides[v.Context.Provides] = append(r.provides[v.Context.Provides], v)
			r.vars[v.satVarName] = v
		}
	}
	logrus.Infof("Loaded %v packages.", len(r.pkgProvides))
	// Generate imply rules
	for _, resourceVars := range r.pkgProvides {
		// Create imply rules for every package and add them to the formula
		// one provided dependency implies all dependencies from that package
		bfVar := bf.And(toBFVars(resourceVars)...)
		var ands []bf.Formula
		for _, res := range resourceVars {
			ands = append(ands, bf.Implies(bf.Var(res.satVarName), bfVar))
		}
		pkgVar := resourceVars[len(resourceVars)-1]
		ands = append(ands, bf.Implies(bf.Var(pkgVar.satVarName), r.explodePackageRequires(pkgVar)))
		if conflicts := r.explodePackageConflicts(pkgVar); conflicts != nil {
			ands = append(ands, bf.Implies(bf.Var(pkgVar.satVarName), bf.Not(conflicts)))
		}
		r.ands = append(r.ands, ands...)
	}
	logrus.Infof("Generated %v variables.", len(r.vars))
	return nil
}

func (r *Resolver) ConstructRequirements(packages []string) error {
	for _, pkgName := range packages {
		req := r.resolveNewest(pkgName)
		logrus.Infof("Selecting %s: %v", pkgName, req.Package)
		r.ands = append(r.ands, bf.Var(req.satVarName))
	}
	return nil
}

func (res *Resolver) Resolve() (install []*api.Package, excluded []*api.Package, err error) {
	logrus.WithField("bf", bf.And(res.ands...)).Debug("Formula to solve")

	if len(res.unresolvable) > 0 {
		return nil, nil, fmt.Errorf("Can't satisfy %+v", res.unresolvable)
	}
	vars := bf.Solve(bf.And(res.ands...))

	if len(vars) > 0 {
		logrus.Info("Solution found.")
		installMap := map[VarContext]*api.Package{}
		excludedMap := map[VarContext]*api.Package{}
		for k, v := range vars {
			resVar := res.vars[k]
			if resVar != nil && resVar.varType == VarTypePackage {
				if v {
					installMap[resVar.Context] = resVar.Package
				} else {
					excludedMap[resVar.Context] = resVar.Package
				}
			}
			//fmt.Printf("%s:%s:%v\n", k, res.vars[k].Context.Provides, v)
		}
		for _, v := range installMap {
			if rpm.Compare(res.bestPackages[v.Name].Version, v.Version) != 0 {
				logrus.Infof("Picking %v instead of best candiate %v", v, res.bestPackages[v.Name])
			}
			install = append(install, v)
		}

		for _, v := range excludedMap {
			excluded = append(excluded, v)
		}
		return install, excluded, nil
	}
	logrus.Info("No solution found.")
	return nil, nil, fmt.Errorf("no solution found")
}

func (res *Resolver) MUS() (mus *explain.Problem, err error) {
	logrus.Info("No solution found.")
	r, w := io.Pipe()

	err = bf.Dimacs(bf.And(res.ands...), w)
	if err != nil {
		return nil, err
	}
	problem, err := explain.ParseCNF(r)
	if err != nil {
		return nil, err
	}
	mus, err = problem.MUSInsertion()
	if err != nil {
		return nil, err
	}
	return mus, nil
}

func (r *Resolver) explodePackageToVars(pkg *api.Package) (pkgVar *Var, resourceVars []*Var) {
	for _, p := range pkg.Format.Provides.Entries {
		if p.Name == pkg.Name {
			pkgVar = &Var{
				satVarName: r.ticket(),
				varType:    VarTypePackage,
				Context: VarContext{
					Package:  pkg.Name,
					Provides: pkg.Name,
					Version:  pkg.Version,
				},
				Package: pkg,
			}
			resourceVars = append(resourceVars, pkgVar)
		} else {
			resVar := &Var{
				satVarName: r.ticket(),
				varType:    VarTypeResource,
				Context: VarContext{
					Package:  pkg.Name,
					Provides: p.Name,
					Version:  pkg.Version,
				},
				Package: pkg,
			}
			resourceVars = append(resourceVars, resVar)
		}
	}

	for _, f := range pkg.Format.Files {
		resVar := &Var{
			satVarName: r.ticket(),
			varType:    VarTypeFile,
			Context: VarContext{
				Package:  pkg.Name,
				Provides: f.Text,
				Version:  pkg.Version,
			},
			Package: pkg,
		}
		resourceVars = append(resourceVars, resVar)
	}
	return pkgVar, resourceVars
}

func (r *Resolver) explodePackageRequires(pkgVar *Var) bf.Formula {
	var bfunique = bf.Var(pkgVar.satVarName)
	for _, req := range pkgVar.Package.Format.Requires.Entries {
		satisfies, err := r.explodeSingleRequires(req, r.provides[req.Name])
		if err != nil {
			r.unresolvable = append(r.unresolvable, req)
			continue
		}
		uniqueVars := []string{}
		for _, s := range satisfies {
			uniqueVars = append(uniqueVars, s.satVarName)
		}
		bfunique = bf.And(bf.Unique(uniqueVars...), bfunique)
	}
	return bfunique
}

func (r *Resolver) explodePackageConflicts(pkgVar *Var) bf.Formula {
	conflictingVars := []bf.Formula{}
	for _, req := range pkgVar.Package.Format.Conflicts.Entries {
		conflicts, err := r.explodeSingleRequires(req, r.provides[req.Name])
		if err != nil {
			// if a conflicting resource does not exist, we don't care
			continue
		}
		for _, s := range conflicts {
			if s.Package == pkgVar.Package {
				// don't conflict with yourself
				continue
			}
			conflictingVars = append(conflictingVars, bf.Var(s.satVarName))
		}
	}
	if len(conflictingVars) == 0 {
		return nil
	}
	return bf.Or(conflictingVars...)
}

func (r *Resolver) resolveNewest(pkgName string) *Var {
	pkgs := r.provides[pkgName]
	newest := pkgs[0]
	for _, p := range pkgs {
		if rpm.Compare(p.Package.Version, newest.Package.Version) == 1 {
			newest = p
		}
	}
	return newest
}

func (r *Resolver) explodeSingleRequires(entry api.Entry, provides []*Var) (accepts []*Var, err error) {
	entryVer := api.Version{
		Text:  entry.Text,
		Epoch: entry.Epoch,
		Ver:   entry.Ver,
		Rel:   entry.Rel,
	}

	for _, dep := range provides {
		cmp := rpm.Compare(dep.Package.Version, entryVer)
		works := false
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
		if works {
			accepts = append(accepts, dep)
		}
	}

	if len(accepts) == 0 {
		return nil, fmt.Errorf("Nothing can satisfy %s", entry.Name)
	}

	return accepts, nil
}
