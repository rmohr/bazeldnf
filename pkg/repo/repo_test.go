package repo

import (
	"fmt"
	"testing"

	"github.com/rmohr/bazeldnf/pkg/api"
)

func Test(t *testing.T) {
	repos, err := LoadRepoFile("repo.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	helper := CacheHelper{CacheDir: ".bazeldnf"}
	a, b, err := helper.CurrentFilelistsForPackages(&repos.Repositories[0], []string{"myarch"}, []*api.Package{
		{Name: "blub", Arch: "myarch", Version: api.Version{Epoch: "1"}},
		{Name: "blub", Arch: "myarch", Version: api.Version{Epoch: "3"}},
		{Name: "blub", Arch: "myarch", Version: api.Version{Epoch: "2"}},
		{Name: "a", Arch: "myarch", Version: api.Version{Epoch: "2"}},
		{Name: "z", Arch: "myarch", Version: api.Version{Epoch: "2"}},
		{Name: "b", Arch: "myarch", Version: api.Version{Epoch: "2"}},
		{Name: "d", Arch: "myarch", Version: api.Version{Epoch: "2"}},
	})
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(err)
}
