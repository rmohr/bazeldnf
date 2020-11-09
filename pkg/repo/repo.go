package repo

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type RepoResolver interface {
	Resolve(out string) error
}

type RepoResolverImpl struct {
	OS                 string
	Arch               string
	PrimaryMetaLinkURL string
	UpdateMetaLinkURL  string
	RepoFile           string
}

func (r RepoResolverImpl) resolveMirror() (*api.File, error) {
	log.Infof("Resolving mirror from %s", r.PrimaryMetaLinkURL)
	resp, err := http.Get(r.PrimaryMetaLinkURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	metalink := &api.Metalink{}
	err = xml.NewDecoder(resp.Body).Decode(metalink)
	if err != nil {
		return nil, err
	}

	var repomod *api.File
	for _, sec := range metalink.Files.File {
		if sec.Name == "repomd.xml" {
			repomod = &sec
			break
		}
	}

	if repomod == nil {
		return nil, fmt.Errorf("Metalink file contains no reference to repod.xml")
	}

	return repomod, nil
}

func (r RepoResolverImpl) resolveRepomd(file *api.File) (repomd *api.Repomd, mirror *url.URL, err error) {
	for _, u := range file.Resources.URLs {
		if u.Protocol != "https" {
			continue
		}
		log.Infof("Resolving repomd.xml from %s", u.Text)
		resp, err := http.Get(u.Text)
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

func (r RepoResolverImpl) resolvePrimary(repomd *api.Repomd, mirror *url.URL) (repoReader io.ReadCloser, err error) {
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
	resp, err := http.Get(primaryURL)
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
	location, err := r.resolveMirror()
	if err != nil {
		return err
	}

	repomd, mirror, err := r.resolveRepomd(location)
	if err != nil {
		return err
	}

	reader, err := r.resolvePrimary(repomd, mirror)
	if err != nil {
		return err
	}
	defer reader.Close()
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("Failed to create output file %s: %v", out, err)
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	if err != nil {
		return fmt.Errorf("Failed to write file to %s: %v", out, err)
	}
	return nil
}

func (r RepoResolverImpl) Init() error {
	repos := &bazeldnf.Repositories{
		Repositories: []bazeldnf.Repository{
			{
				Disabled: false,
				Metalink: r.PrimaryMetaLinkURL,
				Baseurl:  "",
				Arch:     r.Arch,
			},
			{
				Disabled: false,
				Metalink: r.UpdateMetaLinkURL,
				Baseurl:  "",
				Arch:     r.Arch,
			},
		},
	}
	data, err := yaml.Marshal(repos)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(r.RepoFile, data, 0660)
}

func NewRemoteRepoResolver(os string, arch string, repoFile string) RepoResolver {
	return &RepoResolverImpl{
		OS:                 os,
		Arch:               arch,
		RepoFile:           repoFile,
		PrimaryMetaLinkURL: fmt.Sprintf("https://mirrors.fedoraproject.org/metalink?repo=fedora-%s&arch=%s", os, arch),
		UpdateMetaLinkURL:  fmt.Sprintf("https://mirrors.fedoraproject.org/metalink?repo=updates-released-%s&arch=%s", os, arch),
	}
}
