package sat

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/crillab/gophersat/bf"
	"github.com/crillab/gophersat/maxsat"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/rpm"
	"github.com/sirupsen/logrus"
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

func Resolve(model *Model) (install []*api.Package, excluded []*api.Package, forceIgnoredWithDependencies []*api.Package, err error) {
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

type ConversionVars struct {
	satToPkg map[string]string
	pkgToSat map[string]string
}
