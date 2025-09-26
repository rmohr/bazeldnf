package reducer

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

type MockPackageLoader struct {
	packageInfo *packageInfo
}

func (m *MockPackageLoader) Load() (*packageInfo, error) {
	return m.packageInfo, nil
}

func resolve(p *packageInfo, requires, implicitRequires []string, ignoreMissing bool) (matched []string, involved []*api.Package, err error) {
	repoReducer := &RepoReducer{
		implicitRequires: implicitRequires,
		loader:           &MockPackageLoader{packageInfo: p},
	}

	if err := repoReducer.Load(); err != nil {
		return nil, nil, err
	}
	return repoReducer.Resolve(requires, ignoreMissing)

}

func TestReducerZeroPackages(t *testing.T) {
	g := NewGomegaWithT(t)
	matched, involved, err := resolve(&packageInfo{}, []string{}, []string{}, false)

	g.Expect(err).Should(BeNil())
	g.Expect(len(matched)).Should(BeZero())
	g.Expect(len(involved)).Should(BeZero())
}

func TestReducerPackageNotFound(t *testing.T) {
	g := NewGomegaWithT(t)
	matched, involved, err := resolve(&packageInfo{}, []string{"foo"}, []string{}, false)

	g.Expect(err).To(MatchError("Package foo does not exist"))
	g.Expect(len(matched)).Should(BeZero())
	g.Expect(len(involved)).Should(BeZero())
}

func TestReducerImplicitPackageNotFound(t *testing.T) {
	g := NewGomegaWithT(t)
	matched, involved, err := resolve(&packageInfo{}, []string{}, []string{"bar"}, false)

	g.Expect(err).To(MatchError("Package bar does not exist"))
	g.Expect(len(matched)).Should(BeZero())
	g.Expect(len(involved)).Should(BeZero())
}

func TestReducerOnlyImplicitRequires(t *testing.T) {
	g := NewGomegaWithT(t)

	packageInfo := packageInfo{
		packages: withRepository(newPackageList("foo")),
	}

	matched, involved, err := resolve(&packageInfo, []string{}, []string{"foo"}, false)

	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&packageInfo.packages[0]))
}

func TestReducerSingleCandidate(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository(newPackageList("bar"))
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"bar"}, []string{}, false)

	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("bar"))
	g.Expect(involved).Should(ConsistOf(&packages[0]), false)
}

func TestReducerSingleCandidateMissing(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository(newPackageList())
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"bar"}, []string{}, true)

	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(BeEmpty())
	g.Expect(involved).Should(BeEmpty())
}

func TestReducerMultipleCandidates(t *testing.T) {
	g := NewGomegaWithT(t)
	packageNames := []string{"foo", "bar", "baz"}
	packages := withRepository(newPackageList(packageNames...))
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, packageNames, []string{}, false)

	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo", "bar", "baz"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1], &packages[2]))
}

func TestReducerMultipleNameMatch(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository(newPackageList("foo", "foo", "bar"))
	packages[0].Version = api.Version{Epoch: "1"}
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"foo", "bar"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo", "bar"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1], &packages[2]))
}

func TestReducerRequiresMissingProvides(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", []string{"bar"}, nil),
	})
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&packages[0]))
}

func TestReducerRequiresFoundProvides(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", []string{"bar"}, nil),
		newPackage("bar"),
	})
	packageInfo := packageInfo{
		packages: packages,
		provides: map[string][]*api.Package{
			"bar": []*api.Package{&packages[1]},
		},
	}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1]))
}

func TestReducerRequiresFoundMultipleProvides(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", []string{"baz", "bam"}, nil),
		newPackage("baz"),
		newPackage("bam"),
	})

	packageInfo := packageInfo{
		packages: packages,
		provides: map[string][]*api.Package{
			"baz": []*api.Package{&packages[1]},
			"bam": []*api.Package{&packages[2]},
		},
	}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1], &packages[2]))
}

func TestReducerRequiresFoundMultipleProvidesInOne(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", []string{"baz", "bam"}, nil),
		newPackage("baz"),
		newPackage("bam"),
	})

	packageInfo := packageInfo{
		packages: packages,
		provides: map[string][]*api.Package{
			"baz": []*api.Package{&packages[1], &packages[2]},
		},
	}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1], &packages[2]))
}

func TestReducerMultiLevelRequires(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", []string{"baz"}, nil),
		newPackageWithDeps("baz", []string{"bam"}, nil),
		newPackage("bam"),
	})

	packageInfo := packageInfo{
		packages: packages,
		provides: map[string][]*api.Package{
			"baz": []*api.Package{&packages[1]},
			"bam": []*api.Package{&packages[2]},
		},
	}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1], &packages[2]))
}

func TestReducerExcludePinnedDependency(t *testing.T) {
	g := NewGomegaWithT(t)
	pinned := newPackage("bar")
	pinned.Version = api.Version{Epoch: "1"}

	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", []string{"bar"}, nil),
		newPackage("bar"),
	})

	packageInfo := packageInfo{
		packages: packages,
		provides: map[string][]*api.Package{
			"bar": []*api.Package{&packages[1], &pinned},
		},
	}

	matched, involved, err := resolve(&packageInfo, []string{"foo", "bar"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo", "bar"))
	g.Expect(involved).Should(ConsistOf(&packages[0], &packages[1]))
}

func TestInvolvedProvidesIsNotRequiredOrSelf(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", nil, []string{"bar"}),
	})
	expectedPackage := newPackageWithDeps("foo", nil, []string{})
	expectedPackage.Repository = &bazeldnf.Repository{}
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&expectedPackage))
}

func TestInvolvedProvidesIsSelf(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository([]api.Package{
		newPackageWithDeps("foo", nil, []string{"foo"}),
	})
	expectedPackage := newPackageWithDeps("foo", nil, []string{"foo"})
	expectedPackage.Repository = &bazeldnf.Repository{}

	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"foo"}, []string{}, false)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo"))
	g.Expect(involved).Should(ConsistOf(&expectedPackage))
}

func TestRepositoryPriority(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository(newPackageList("bar", "bar"))
	packages[0].Repository.Priority = 2
	packages[1].Repository.Priority = 1
	packages[1].Summary = "I'm the one"

	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"bar"}, []string{}, false)

	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("bar"))
	g.Expect(involved).Should(ConsistOf(&packages[1]))
}

func TestRepositoryPriorityWithVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository(newPackageList("bar", "bar", "bar"))
	packages[0].Repository.Priority = 2
	packages[1].Repository.Priority = 1
	packages[1].Summary = "I'm the one"
	packages[2].Version = api.Version{Epoch: "3"}

	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"bar"}, []string{}, false)

	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("bar"))
	g.Expect(involved).Should(ConsistOf(&packages[1], &packages[2]))
}

func TestSpecifyVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	packages := withRepository(newPackageList("foo", "foo", "foo", "bar", "baz", "baz"))
	packages[0].Version = api.Version{Epoch: "1", Ver: "3", Rel: "4"}
	packages[1].Version = api.Version{Epoch: "2", Ver: "3", Rel: "4"}
	packages[2].Version = api.Version{Epoch: "2", Ver: "3", Rel: "5"}
	packages[3].Version = api.Version{Epoch: "1", Ver: "9", Rel: "8"}
	packages[4].Version = api.Version{Epoch: "1", Ver: "1.13", Rel: "1"}
	packages[5].Version = api.Version{Epoch: "1", Ver: "1.14", Rel: "6"}
	packageInfo := packageInfo{packages: packages}

	matched, involved, err := resolve(&packageInfo, []string{"foo-2:3", "bar-2", "baz-1:1.13-1"}, []string{}, true)
	g.Expect(err).Should(BeNil())
	g.Expect(matched).Should(ConsistOf("foo", "baz"))
	g.Expect(involved).Should(ConsistOf(&packages[1], &packages[2], &packages[4]))
}
