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
	a, b, err := helper.CurrentFilelistsForPackages(&repos.Repositories[0], []*api.Package{
		{Name: "blub", Version: api.Version{Epoch: "1"}},
		{Name: "blub", Version: api.Version{Epoch: "3"}},
		{Name: "blub", Version: api.Version{Epoch: "2"}},
		{Name: "a", Version: api.Version{Epoch: "2"}},
		{Name: "z", Version: api.Version{Epoch: "2"}},
		{Name: "b", Version: api.Version{Epoch: "2"}},
		{Name: "d", Version: api.Version{Epoch: "2"}},
	})
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(err)
}
