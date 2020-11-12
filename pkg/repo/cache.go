package repo

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
	"github.com/rmohr/bazeldnf/pkg/rpm"
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
	primary := repomd.File(api.PrimaryFileType)
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

func (r *CacheHelper) CurrentFilelistsForPackages(repo *bazeldnf.Repository, packages []*api.Package) (filelistpkgs []*api.FileListPackage, remaining []*api.Package, err error) {
	repomd := &api.Repomd{}
	if err := r.UnmarshalFromRepoDir(repo, "repomd.xml", repomd); err != nil {
		return nil, nil, err
	}
	filelists := repomd.File(api.FilelistsFileType)
	filelistsName := filepath.Base(filelists.Location.Href)
	file, err := r.OpenFromRepoDir(repo, filelistsName)
	if err != nil {
		return nil, nil, err
	}

	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	d := xml.NewDecoder(reader)
	pkgIndex := 0

	sort.SliceStable(packages, func(i, j int) bool {
		if packages[i].Name == packages[j].Name {
			return rpm.Compare(packages[i].Version, packages[j].Version) < 0
		}
		return packages[i].Name < packages[j].Name
	})

	for {
		if len(packages) == pkgIndex {
			break
		}
		currPkg := packages[pkgIndex]
		tok, err := d.Token()
		if tok == nil || err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, fmt.Errorf("Error decoding token: %s", err)
		}

		switch ty := tok.(type) {
		case xml.StartElement:
			if ty.Name.Local == "package" {
				name := ""
				for _, attr := range ty.Attr {
					if attr.Name.Local == "name" {
						name = attr.Value
					}
				}
				if currPkg.Name == name {
					pkg := &api.FileListPackage{}
					if err = d.DecodeElement(pkg, &ty); err != nil {
						return nil, nil, fmt.Errorf("Error decoding item: %s", err)
					}
					if rpm.Compare(currPkg.Version, pkg.Version) == 0 {
						pkgIndex++
						filelistpkgs = append(filelistpkgs, pkg)
					}
				} else if name > currPkg.Name {
					remaining = append(remaining, currPkg)
					pkgIndex++
				}
			}
		default:
		}
	}

	return filelistpkgs, remaining, nil
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
