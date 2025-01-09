package reducer

import (
	"github.com/rmohr/bazeldnf/pkg/api"
)

func newPackageList(names ...string) []api.Package {
	r := []api.Package{}
	for _, name := range names {
		r = append(r, newPackage(name))
	}
	return r
}

func newPackage(name string) api.Package {
	r := api.Package{
		Name: name,
		Arch: "x86_64",
	}
	r.Format.Requires = api.Dependencies{Entries: []api.Entry{}}

	return r
}

func toDeps(deps ...string) api.Dependencies {
	e := []api.Entry{}
	if deps == nil {
		e = nil
	} else {
		for _, dep := range deps {
			e = append(e, api.Entry{Name: dep})
		}
	}
	return api.Dependencies{Entries: e}
}

func newPackageWithDeps(name string, requires, provides []string) api.Package {
	p := newPackage(name)

	p.Format.Requires = toDeps(requires...)
	p.Format.Provides = toDeps(provides...)
	return p
}
