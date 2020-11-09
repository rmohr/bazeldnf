package repo

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	log "github.com/sirupsen/logrus"
)

type Getter interface {
	Get(url string) (resp *http.Response, err error)
}

type getterImpl struct{}

func (*getterImpl) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

type RepoResolver interface {
	Resolve(out string) error
}

type RepoResolverImpl struct {
	Getter      Getter
	Repos       []bazeldnf.Repository
	CacheHelper *CacheHelper
}

func (r RepoResolverImpl) resolveRepomd(file *api.File) (repomd *api.Repomd, mirror *url.URL, err error) {
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
		file := &api.Repomd{}
		err = xml.NewDecoder(resp.Body).Decode(file)
		if err != nil {
			log.Errorf("Failed to decode repomd.xml from %s: %v", u.Text, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
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

func (r RepoResolverImpl) fetchRepoXML(repomd *api.Repomd, mirror *url.URL) (repoReader io.ReadCloser, err error) {
	var primary *api.Data
	for _, data := range repomd.Data {
		if data.Type == "primary" {
			primary = &data
			break
		}
	}
	if primary == nil {
		return nil, fmt.Errorf("No 'primary' file referenced in repomd")
	}
	if primary.Location.Href == "" {
		return nil, fmt.Errorf("The 'primary' file has no href associated")
	}

	primaryURL := primary.Location.Href
	if !path.IsAbs(primary.Location.Href) {
		mirrorCopy := *mirror
		mirrorCopy.Path = path.Join(mirror.Path, primary.Location.Href)
		primaryURL = mirrorCopy.String()
	}
	log.Infof("Loading primary repository file from %s", primaryURL)
	resp, err := r.Getter.Get(primaryURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to load promary repository file from %s: %v", primaryURL, err)
	}
	reader := resp.Body
	if resp.Header.Get("Content-Type") == "application/x-gzip" {
		reader, err = gzip.NewReader(resp.Body)
	}
	return reader, nil
}

func (r RepoResolverImpl) Resolve(out string) error {
	return nil
}

func NewRemoteRepoResolver(repos []bazeldnf.Repository, cacheDir string) RepoResolver {
	return &RepoResolverImpl{
		Repos:    repos,
		Getter:   &getterImpl{},
		CacheHelper: &CacheHelper{CacheDir: cacheDir},
	}
}
