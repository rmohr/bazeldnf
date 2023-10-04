package repo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"
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

func (r *RepoFetcherImpl) Fetch() (err error) {
	for _, repo := range r.Repos {
		sha256sum := []string{}
		var repomdURLs = []string{}
		if repo.Metalink != "" {
			var metalink *api.Metalink
			metalink, repomdURLs, err = r.resolveMetaLink(&repo)
			if err != nil {
				return fmt.Errorf("failed to resolve metalink for %s: %v", repo.Name, err)
			}
			sha256sum, err = metalink.Repomod().SHA256()
			if err != nil {
				return fmt.Errorf("failed to get sha256sum of repomd file: %v", err)
			}
		} else if repo.Baseurl != "" {
			repomdURLs = append(repomdURLs, strings.TrimSuffix(repo.Baseurl, "/")+"/repodata/repomd.xml")
		}
		repomd, mirror, err := r.resolveRepomd(&repo, repomdURLs, sha256sum)
		if err != nil {
			return fmt.Errorf("failed to fetch repomd.xml for %s: %v", repo.Name, err)
		}
		err = r.fetchFile(api.PrimaryFileType, &repo, repomd, mirror)
		if err != nil {
			return fmt.Errorf("failed to fetch primary.xml for %s: %v", repo.Name, err)
		}
		/* not used right now, save some bandwidth
		err = r.fetchFile(api.FilelistsFileType, &repo, repomd, mirror)
		if err != nil {
			return fmt.Errorf("failed to fetch filelists.xml for %s: %v", repo.Name, err)
		}
		*/
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

func (r *RepoFetcherImpl) resolveMetaLink(repo *bazeldnf.Repository) (*api.Metalink, []string, error) {
	resp, err := r.Getter.Get(repo.Metalink)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, nil, fmt.Errorf("Failed to download %s: %v ", repo.Metalink, fmt.Errorf("status : %v", resp.StatusCode))
	}
	if err := r.CacheHelper.WriteToRepoDir(repo, resp.Body, "metalink"); err != nil {
		return nil, nil, err
	}

	metalink, err := r.CacheHelper.LoadMetaLink(repo)
	if err != nil {
		return nil, nil, err
	}

	repomod := metalink.Repomod()

	if repomod == nil {
		return nil, nil, fmt.Errorf("Metalink file contains no reference to repod.xml")
	}

	urls := []string{}
	for _, u := range repomod.Resources.URLs {
		if u.Protocol != "https" {
			continue
		}
		urls = append(urls, u.Text)
	}

	if len(urls) == 0 {
		return metalink, nil, fmt.Errorf("Metalink contains no https url to a rpomd.xml file")
	}

	return metalink, urls, nil
}

func (r *RepoFetcherImpl) resolveRepomd(repo *bazeldnf.Repository, repomdURLs []string, sha256sums []string) (repomd *api.Repomd, mirror *url.URL, err error) {
	for _, u := range repomdURLs {
		sha := sha256.New()
		log.Infof("Resolving repomd.xml from %s", u)
		resp, err := r.Getter.Get(u)
		if err != nil {
			log.Errorf("Failed to resolve repomd.xml from %s: %v", u, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			log.Warningf("Failed to download %s: %v ", u, fmt.Errorf("status : %v", resp.StatusCode))
			continue
		}
		body := io.TeeReader(resp.Body, sha)
		err = r.CacheHelper.WriteToRepoDir(repo, body, "repomd.xml")
		if err != nil {
			log.Errorf("Failed to save repomd.xml from %s: %v", u, err)
			continue
		}
		if len(sha256sums) > 0 {
			matched := false
			for _, sum := range sha256sums {
				if toHex(sha) != sum {
					log.Warnf("Expected repomd.xml sha256 sum %s, but got %s", sum, toHex(sha))
				} else {
					log.Infof("Matched repmod.xml with sha256 sum %s", toHex(sha))
					matched = true
					break
				}
			}
			if !matched {
				log.Warningf("Mirror has no expected repomd.xml version: %v", u)
				continue
			}
		}

		file := &api.Repomd{}
		err = r.CacheHelper.UnmarshalFromRepoDir(repo, "repomd.xml", file)
		if err != nil {
			log.Errorf("Failed to decode repomd.xml from %s: %v", u, err)
			continue
		}
		repomd = file
		mirror, err = url.Parse(u)
		if err != nil {
			log.Fatalf("Invalid URL for repomd.xml from %s, this should be impossible: %v", u, err)
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
		return fmt.Errorf("Failed to load primary repository file from %s: %v", fileURL, err)
	}
	sha := sha256.New()
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Failed to download %s: %v ", fileURL, fmt.Errorf("status : %v", resp.StatusCode))
	}
	body := io.TeeReader(resp.Body, sha)
	err = r.CacheHelper.WriteToRepoDir(repo, body, fileName)
	if err != nil {
		return fmt.Errorf("Failed to write file.xml from %s to file: %v", fileURL, err)
	}
	sha256sum, err := file.SHA256()
	if err != nil {
		return fmt.Errorf("failed to get sha256sum of file: %v", err)
	}
	if sha256sum != toHex(sha) {
		return fmt.Errorf("Expected sha256 sum %s, but got %s", sha256sum, toHex(sha))
	}
	return nil
}

type Getter interface {
	Get(url string) (resp *http.Response, err error)
}

type getterImpl struct{}

func fileGet(filename string) (*http.Response, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err // skipped wrapping the error since the error already begins with "open: "
	}

	resp := &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       fp,
	}
	return resp, nil
}

func (*getterImpl) Get(rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL: %w", err)
	}
	if u.Scheme == "file" {
		return fileGet(u.Path)
	}
	return http.Get(rawURL)
}

func toHex(hasher hash.Hash) string {
	return hex.EncodeToString(hasher.Sum(nil))
}
