package api

import (
	"encoding/xml"
	"fmt"

	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

const (
	PrimaryFileType   = "primary"
	FilelistsFileType = "filelists"
)

type URL struct {
	Text       string `xml:",chardata"`
	Protocol   string `xml:"protocol,attr"`
	Type       string `xml:"type,attr"`
	Location   string `xml:"location,attr"`
	Preference string `xml:"preference,attr"`
}

type File struct {
	Name      string `xml:"name,attr"`
	Resources struct {
		URLs []URL `xml:"url"`
	} `xml:"resources"`
	Timestamp    string `xml:"timestamp"`
	Size         string `xml:"size"`
	Verification struct {
		Hash []struct {
			Hash string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"hash"`
	} `xml:"verification"`
	Alternates struct {
		Text      string        `xml:",chardata"`
		Alternate []Alternative `xml:"alternate"`
	} `xml:"alternates"`
}

type Alternative struct {
	Text         string `xml:",chardata"`
	Timestamp    string `xml:"timestamp"`
	Size         string `xml:"size"`
	Verification struct {
		Text string `xml:",chardata"`
		Hash []struct {
			Hash string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"hash"`
	} `xml:"verification"`
}

func (f *File) SHA256() (sums []string, err error) {
	for _, h := range f.Verification.Hash {
		if h.Type == "sha256" {
			sums = append(sums, h.Hash)
		}
	}
	for _, a := range f.Alternates.Alternate {
		for _, h := range a.Verification.Hash {
			if h.Type == "sha256" {
				sums = append(sums, h.Hash)
			}
		}
	}
	if len(sums) == 0 {
		return nil, fmt.Errorf("no sha256 found")
	}
	return
}

type Metalink struct {
	XMLName xml.Name `xml:"metalink"`
	Files   struct {
		File []File ` xml:"file"`
	} `xml:"files"`
}

func (m *Metalink) Repomod() *File {
	var repomod *File
	for _, sec := range m.Files.File {
		if sec.Name == "repomd.xml" {
			repomod = &sec
			break
		}
	}
	return repomod
}

type Data struct {
	Text     string `xml:",chardata"`
	Type     string `xml:"type,attr"`
	Checksum struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
	} `xml:"checksum"`
	OpenChecksum struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
	} `xml:"open-checksum"`
	Location struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
	} `xml:"location"`
	Timestamp       string `xml:"timestamp"`
	Size            string `xml:"size"`
	OpenSize        string `xml:"open-size"`
	DatabaseVersion string `xml:"database_version"`
	HeaderChecksum  struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
	} `xml:"header-checksum"`
	HeaderSize string `xml:"header-size"`
}

func (d *Data) SHA512() (string, error) {
	if d.Checksum.Type == "sha512" {
		return d.Checksum.Text, nil
	}
	return "", fmt.Errorf("no sha512 found")
}

func (d *Data) SHA256() (string, error) {
	if d.Checksum.Type == "sha256" {
		return d.Checksum.Text, nil
	}
	return "", fmt.Errorf("no sha256 found")
}

func (d *Data) SHA() (string, error) {
	if d.Checksum.Type == "sha" {
		return d.Checksum.Text, nil
	}
	return "", fmt.Errorf("no sha found")
}

type Repomd struct {
	XMLName  xml.Name `xml:"repomd"`
	Text     string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	Rpm      string   `xml:"rpm,attr"`
	Revision string   `xml:"revision"`
	Data     []Data   `xml:"data"`
}

func (r *Repomd) File(fileType string) *Data {
	var primary *Data
	for _, data := range r.Data {
		if data.Type == fileType {
			primary = &data
			break
		}
	}
	return primary
}

func (r *Repomd) Filelists() *Data {
	var primary *Data
	for _, data := range r.Data {
		if data.Type == "filelists" {
			primary = &data
			break
		}
	}
	return primary
}

type Entry struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name,attr"`
	Flags string `xml:"flags,attr"`
	Epoch string `xml:"epoch,attr"`
	Ver   string `xml:"ver,attr"`
	Rel   string `xml:"rel,attr"`
}

func (e Entry) String() string {
	if e.Flags != "" {
		v := Version{
			Epoch: e.Epoch,
			Ver:   e.Ver,
			Rel:   e.Rel,
		}
		return fmt.Sprintf("%s-%s-%s", e.Name, e.Flags, v.String())
	}
	return e.Name
}

type Dependencies struct {
	Entries []Entry `xml:"entry"`
}

type ProvidedFile struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

type Version struct {
	Text  string `xml:",chardata"`
	Epoch string `xml:"epoch,attr"`
	Ver   string `xml:"ver,attr"`
	Rel   string `xml:"rel,attr"`
}

func (v *Version) String() string {
	var version string
	if v.Epoch == "" {
		version = "0"
	} else {
		version = v.Epoch
	}
	version = version + ":" + v.Ver

	if v.Rel != "" {
		version = version + "-" + v.Rel
	}
	return version
}

type Checksum struct {
	Text  string `xml:",chardata"`
	Type  string `xml:"type,attr"`
	Pkgid string `xml:"pkgid,attr"`
}

func (c Checksum) Integrity() (string, error) {
	if c.Type == "sha" {
		return fmt.Sprintf("sha1-%s", c.Text), nil
	} else if c.Type == "sha512" || c.Type == "sha256" {
		return fmt.Sprintf("%s-%s", c.Type, c.Text), nil
	}

	return "", fmt.Errorf("Invalid integrity type: %s", c.Type)
}

type Location struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
}

type Package struct {
	Type        string   `xml:"type,attr"`
	Name        string   `xml:"name"`
	Arch        string   `xml:"arch"`
	Version     Version  `xml:"version"`
	Checksum    Checksum `xml:"checksum"`
	Summary     string   `xml:"summary"`
	Description string   `xml:"description"`
	Packager    string   `xml:"packager"`
	URL         string   `xml:"url"`
	Time        struct {
		Text  string `xml:",chardata"`
		File  string `xml:"file,attr"`
		Build string `xml:"build,attr"`
	} `xml:"time"`
	Size struct {
		Text      string `xml:",chardata"`
		Package   int    `xml:"package,attr"`
		Installed int    `xml:"installed,attr"`
		Archive   int    `xml:"archive,attr"`
	} `xml:"size"`
	Location Location `xml:"location"`
	Format   struct {
		License     string `xml:"license"`
		Vendor      string `xml:"vendor"`
		Group       string `xml:"group"`
		Buildhost   string `xml:"buildhost"`
		Sourcerpm   string `xml:"sourcerpm"`
		HeaderRange struct {
			Text  string `xml:",chardata"`
			Start string `xml:"start,attr"`
			End   string `xml:"end,attr"`
		} `xml:"header-range"`
		Provides    Dependencies   `xml:"provides"`
		Requires    Dependencies   `xml:"requires"`
		Files       []ProvidedFile `xml:"file"`
		Conflicts   Dependencies   `xml:"conflicts"`
		Obsoletes   Dependencies   `xml:"obsoletes"`
		Recommends  Dependencies   `xml:"recommends"`
		Suggests    Dependencies   `xml:"suggests"`
		Enhances    Dependencies   `xml:"enhances"`
		Supplements Dependencies   `xml:"supplements"`
	} `xml:"format"`
	Repository *bazeldnf.Repository `xml:"-"`
}

func (p *Package) String() string {
	return p.Name + "-" + p.Version.String()
}

type Repository struct {
	XMLName      xml.Name  `xml:"metadata"`
	Text         string    `xml:",chardata"`
	Xmlns        string    `xml:"xmlns,attr"`
	Rpm          string    `xml:"rpm,attr"`
	PackageCount string    `xml:"packages,attr"`
	Packages     []Package `xml:"package"`
}

type Filelists struct {
	XMLName  xml.Name `xml:"filelists"`
	Text     string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	Packages string   `xml:"packages,attr"`
	Package  []struct {
		Text    string `xml:",chardata"`
		Pkgid   string `xml:"pkgid,attr"`
		Name    string `xml:"name,attr"`
		Arch    string `xml:"arch,attr"`
		Version struct {
			Text  string `xml:",chardata"`
			Epoch string `xml:"epoch,attr"`
			Ver   string `xml:"ver,attr"`
			Rel   string `xml:"rel,attr"`
		} `xml:"version"`
		File []struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"file"`
	} `xml:"package"`
}

type FileListPackage struct {
	Text    string         `xml:",chardata"`
	Pkgid   string         `xml:"pkgid,attr"`
	Name    string         `xml:"name,attr"`
	Arch    string         `xml:"arch,attr"`
	Version Version        `xml:"version"`
	File    []ProvidedFile `xml:"file"`
}

func (p *FileListPackage) String() string {
	return p.Name + "-" + p.Version.String()
}
