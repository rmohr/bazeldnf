package repo

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

type CacheHelper struct {
	CacheDir string
}

func (r *CacheHelper) WriteToRepoDir(repo *bazeldnf.Repository, body io.Reader, name string) error {
	dir := filepath.Join(r.CacheDir, repo.Name)
	file := filepath.Join(dir, name)

	err := os.MkdirAll(dir, 0770)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create cache directory for %s: %v", repo.Name, err)
	}
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", file, err)
	}
	defer f.Close()
	_, err = io.Copy(f, body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", file, err)
	}
	return nil
}

func (r *CacheHelper) OpenFromRepoDir(repo *bazeldnf.Repository, name string) (io.ReadCloser, error) {
	dir := filepath.Join(r.CacheDir, repo.Name)
	file := filepath.Join(dir, name)
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", file, err)
	}
	return f, err
}

func (r *CacheHelper) UnmarshalFromRepoDir(repo *bazeldnf.Repository, name string, obj interface{}) error {
	reader, err := r.OpenFromRepoDir(repo, name)
	if err != nil {
		return err
	}
	defer reader.Close()
	return xml.NewDecoder(reader).Decode(obj)
}

func (r *CacheHelper) CurrentPrimary(repo *bazeldnf.Repository) (*api.Repository, error) {
	repomd := &api.Repomd{}
	if err := r.UnmarshalFromRepoDir(repo, "repomd.xml", repomd); err != nil {
		return nil, err
	}
	primary := repomd.Primary()
	primaryName := filepath.Base(primary.Location.Href)
	file, err := r.OpenFromRepoDir(repo, primaryName)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	repository := &api.Repository{}
	err = xml.NewDecoder(reader).Decode(repository)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func (r *CacheHelper) CurrentPrimaries(repos *bazeldnf.Repositories) (primaries []*api.Repository, err error) {
	for _, repo := range repos.Repositories {
		primary, err := r.CurrentPrimary(&repo)
		if err != nil {
			return nil, err
		}
		primaries = append(primaries, primary)
	}
	return primaries, err
}