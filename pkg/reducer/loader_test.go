package reducer

import (
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

type ErrorCacheHelper struct {
	err error
}

func (h ErrorCacheHelper) CurrentPrimaries(_ *bazeldnf.Repositories, _ string) (primaries []*api.Repository, err error) {
	return nil, h.err
}

type MockCacheHelper struct {
	repos []*api.Repository
}

func (h MockCacheHelper) CurrentPrimaries(_ *bazeldnf.Repositories, _ string) (primaries []*api.Repository, err error) {
	return h.repos, nil
}

func load(t *testing.T, repos []api.Repository, architectures []string, cacheHelper RepoCache) (*packageInfo, error) {
	tempdir := t.TempDir()
	repoFiles := []string{}

	for n, repo := range repos {
		repoFile := path.Join(tempdir, fmt.Sprintf("repo%d.xml", n))

		f, err := os.OpenFile(repoFile, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if err := xml.NewEncoder(f).Encode(&repo); err != nil {
			panic(err)
		}

		repoFiles = append(repoFiles, repoFile)
	}

	repoReducer := &RepoLoader{
		repoFiles:     repoFiles,
		architectures: architectures,
		cacheHelper:   cacheHelper,
	}

	return repoReducer.Load()
}

func TestLoaderZeroPackages(t *testing.T) {
	g := NewGomegaWithT(t)
	packageInfo, err := load(t, []api.Repository{}, []string{}, MockCacheHelper{})

	g.Expect(err).Should(BeNil())
	g.Expect(len(packageInfo.packages)).Should(BeZero())
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderOneRepoFileNoPackages(t *testing.T) {
	g := NewGomegaWithT(t)
	packageInfo, err := load(t, []api.Repository{api.Repository{}}, []string{}, MockCacheHelper{})

	g.Expect(err).Should(BeNil())
	g.Expect(len(packageInfo.packages)).Should(BeZero())
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderOneRepoOnePackageSkipOnArch(t *testing.T) {
	g := NewGomegaWithT(t)
	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{
				Packages: newPackageList("mypackage"),
			},
		},
		[]string{"aarch64"},
		MockCacheHelper{},
	)

	g.Expect(err).Should(BeNil())
	g.Expect(len(packageInfo.packages)).Should(BeZero())
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderOneRepoOnePackage(t *testing.T) {
	g := NewGomegaWithT(t)

	packages := newPackageList("mypackage")
	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: packages},
		},
		[]string{"x86_64"},
		MockCacheHelper{},
	)

	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(packages))
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderOneRepoFileManyPackages(t *testing.T) {
	g := NewGomegaWithT(t)

	packages := newPackageList("foo", "bar")
	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: packages},
		},
		[]string{"x86_64"},
		MockCacheHelper{},
	)

	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(packages))
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderManyRepoManyPackages(t *testing.T) {
	g := NewGomegaWithT(t)

	repoAPackages := newPackageList("foo", "bar")
	repoBPackages := newPackageList("baz", "bam")
	repoBPackages[1].Arch = "aarch64"

	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: repoAPackages},
			api.Repository{Packages: repoBPackages},
		},
		[]string{"x86_64"},
		MockCacheHelper{},
	)

	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(append(repoAPackages, repoBPackages[0])))
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderFixPackages(t *testing.T) {
	g := NewGomegaWithT(t)
	dep := "/usr/libexec/platform-python"

	packages := []api.Package{
		newPackageWithDeps("platform-python", []string{dep}, nil),
	}

	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: packages},
		},
		[]string{"x86_64"},
		MockCacheHelper{},
	)

	expectedPackages := newPackageWithDeps("platform-python", []string{}, []string{dep})
	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(expectedPackages))
	g.Expect(packageInfo.provides).Should(BeEquivalentTo(
		map[string][]*api.Package{
			"/usr/libexec/platform-python": []*api.Package{&expectedPackages},
		},
	))
}

func TestLoaderCurrentPrimariesError(t *testing.T) {
	g := NewGomegaWithT(t)
	myErr := fmt.Errorf("My error")
	packageInfo, err := load(t, []api.Repository{}, []string{}, ErrorCacheHelper{err: myErr})

	g.Expect(err).Should(MatchError(err))
	g.Expect(len(packageInfo.packages)).Should(BeZero())
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderHasCachedPrimaries(t *testing.T) {
	g := NewGomegaWithT(t)

	packages := newPackageList("foo", "bar")
	packageInfo, err := load(
		t,
		[]api.Repository{},
		[]string{"x86_64"},
		MockCacheHelper{
			repos: []*api.Repository{
				&api.Repository{Packages: packages},
			},
		},
	)

	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(append(packages)))
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderCachedVsRealRepo(t *testing.T) {
	g := NewGomegaWithT(t)

	repoAPackages := newPackageList("foo", "bar")
	repoBPackages := newPackageList("baz", "bam")
	repoBPackages[1].Arch = "aarch64"

	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: repoAPackages},
		},
		[]string{"x86_64"},
		MockCacheHelper{
			repos: []*api.Repository{
				&api.Repository{Packages: repoBPackages},
			},
		},
	)

	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(append(repoAPackages, repoBPackages[0])))
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderPruneRequires(t *testing.T) {
	g := NewGomegaWithT(t)

	repoPackages := []api.Package{
		newPackageWithDeps("baf", []string{"burgle", "(burgle)"}, nil),
	}

	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: repoPackages},
		},
		[]string{"x86_64"},
		MockCacheHelper{},
	)

	expectedPackages := newPackageWithDeps("baf", []string{"burgle"}, nil)

	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(expectedPackages))
	g.Expect(len(packageInfo.provides)).Should(BeZero())
}

func TestLoaderCaptureProvides(t *testing.T) {
	g := NewGomegaWithT(t)

	repoPackages := []api.Package{
		newPackageWithDeps("baf", []string{}, []string{"burgle"}),
	}
	repoPackages[0].Format.Files = []api.ProvidedFile{
		api.ProvidedFile{Text: "bazzle"},
	}

	packageInfo, err := load(
		t,
		[]api.Repository{
			api.Repository{Packages: repoPackages},
		},
		[]string{"x86_64"},
		MockCacheHelper{},
	)

	expectedProvides := map[string][]*api.Package{
		"burgle": []*api.Package{&repoPackages[0]},
		"bazzle": []*api.Package{&repoPackages[0]},
	}
	g.Expect(err).Should(BeNil())
	g.Expect(packageInfo.packages).Should(ConsistOf(repoPackages[0]))
	g.Expect(packageInfo.provides).Should(BeComparableTo(expectedProvides))
}
