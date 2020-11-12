package repo

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	log "github.com/sirupsen/logrus"
)

type RepoFetcher interface {
	Fetch() error
}

type RepoFetcherImpl struct {
	Getter      Getter
	Repos       []bazeldnf.Repository
	CacheHelper *CacheHelper
}

func (r *RepoFetcherImpl) Fetch() error {
	for _, repo := range r.Repos {
		metalink, err := r.resolveMetaLink(&repo)
		if err != nil {
			return fmt.Errorf("failed to resolve metalink for %s: %v", repo.Name, err)
		}
		repomd, mirror, err := r.resolveRepomd(&repo, metalink.Repomod())
		if err != nil {
			return fmt.Errorf("failed to fetch repomd.xml for %s: %v", repo.Name, err)
		}
		err = r.fetchFile(api.PrimaryFileType, &repo, repomd, mirror)
		if err != nil {
			return fmt.Errorf("failed to fetch primary.xml for %s: %v", repo.Name, err)
		}
		err = r.fetchFile(api.FilelistsFileType, &repo, repomd, mirror)
		if err != nil {
			return fmt.Errorf("failed to fetch filelists.xml for %s: %v", repo.Name, err)
		}
	}
	return nil
}

func NewRemoteRepoFetcher(repos []bazeldnf.Repository, cacheDir string) RepoFetcher {
	return &RepoFetcherImpl{
		Repos:       repos,
		Getter:      &getterImpl{},
		CacheHelper: &CacheHelper{CacheDir: cacheDir},
	}
}

func (r *RepoFetcherImpl) resolveMetaLink(repo *bazeldnf.Repository) (*api.Metalink, error) {
	resp, err := r.Getter.Get(repo.Metalink)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := r.CacheHelper.WriteToRepoDir(repo, resp.Body, "metalink"); err != nil {
		return nil, err
	}

	metalink := &api.Metalink{}
	if err := r.CacheHelper.UnmarshalFromRepoDir(repo, "metalink", metalink); err != nil {
		return nil, err
	}

	repomod := metalink.Repomod()

	if repomod == nil {
		return nil, fmt.Errorf("Metalink file contains no reference to repod.xml")
	}

	return metalink, nil
}

func (r *RepoFetcherImpl) resolveRepomd(repo *bazeldnf.Repository, file *api.File) (repomd *api.Repomd, mirror *url.URL, err error) {
	for _, u := range file.Resources.URLs {
		if u.Protocol != "https" {
			continue
		}
		log.Infof("Resolving repomd.xml from %s", u.Text)
		resp, err := r.Getter.Get(u.Text)
		if err != nil {
			log.Errorf("Failed to resolve repomd.xml from %s: %v", u.Text, err)
			continue
		}
		defer resp.Body.Close()
		err = r.CacheHelper.WriteToRepoDir(repo, resp.Body, "repomd.xml")
		if err != nil {
			log.Errorf("Failed to save repomd.xml from %s: %v", u.Text, err)
			continue
		}
		file := &api.Repomd{}
		err = r.CacheHelper.UnmarshalFromRepoDir(repo, "repomd.xml", file)
		if err != nil {
			log.Errorf("Failed to decode repomd.xml from %s: %v", u.Text, err)
			continue
		}
		repomd = file
		mirror, err = url.Parse(u.Text)
		if err != nil {
			log.Fatalf("Invalid URL for repomd.xml from %s, this should be impossible: %v", u.Text, err)
		}
		break
	}

	if repomd == nil {
		return nil, nil, fmt.Errorf("All mirrors tried, could not download repomd.xml")
	}
	mirror.Path = strings.TrimSuffix(path.Dir(mirror.Path), "repodata")
	return repomd, mirror, nil
}

func (r *RepoFetcherImpl) fetchFile(fileType string, repo *bazeldnf.Repository, repomd *api.Repomd, mirror *url.URL) (err error) {
	file := repomd.File(fileType)
	if file == nil {
		return fmt.Errorf("No 'file' file referenced in repomd")
	}
	if file.Location.Href == "" {
		return fmt.Errorf("The 'file' file has no href associated")
	}

	fileURL := file.Location.Href
	fileName := filepath.Base(file.Location.Href)
	if !path.IsAbs(file.Location.Href) {
		mirrorCopy := *mirror
		mirrorCopy.Path = path.Join(mirror.Path, file.Location.Href)
		fileURL = mirrorCopy.String()
	}
	log.Infof("Loading %s file from %s", fileType, fileURL)
	resp, err := r.Getter.Get(fileURL)
	if err != nil {
		return fmt.Errorf("Failed to load promary repository file from %s: %v", fileURL, err)
	}
	defer resp.Body.Close()
	err = r.CacheHelper.WriteToRepoDir(repo, resp.Body, fileName)
	if err != nil {
		return fmt.Errorf("Failed to write file.xml from %s to file: %v", fileURL, err)
	}
	return nil
}

type Getter interface {
	Get(url string) (resp *http.Response, err error)
}

type getterImpl struct{}

func (*getterImpl) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}