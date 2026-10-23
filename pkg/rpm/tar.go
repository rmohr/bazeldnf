package rpm

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sassoftware/go-rpmutils"
	"github.com/sassoftware/go-rpmutils/cpio"
	log "github.com/sirupsen/logrus"
)

type Collector struct {
	createdPaths map[string]struct{}
}

func NewCollector() *Collector {
	return &Collector{
		createdPaths: make(map[string]struct{}),
	}
}

func (c *Collector) RPMToTar(rpmReader io.Reader, tarWriter *tar.Writer, noSymlinksAndDirs bool, capabilities map[string][]string, selinuxLabels map[string]string) error {
	rpm, err := rpmutils.ReadRpm(rpmReader)
	if err != nil {
		return fmt.Errorf("failed to read rpm: %s", err)
	}
	payloadReader, err := rpm.RawUncompressedRPMPayloadReader()
	if err != nil {
		return fmt.Errorf("failed to open the payload reader: %s", err)
	}
	return Tar(payloadReader, tarWriter, noSymlinksAndDirs, capabilities, selinuxLabels, c.createdPaths)
}

func RPMToCPIO(rpmReader io.Reader) (*cpio.CpioStream, error) {
	rpm, err := rpmutils.ReadRpm(rpmReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read rpm: %s", err)
	}
	payloadReader, err := rpm.RawUncompressedRPMPayloadReader()
	if err != nil {
		return nil, fmt.Errorf("failed to open the payload reader: %s", err)
	}
	return cpio.NewCpioStream(payloadReader), nil
}

var usrMergeDirs = []string{"/bin", "/sbin", "/lib", "/lib64"}

func isUsrMergeRoot(p string) bool {
	cleaned := strings.TrimSuffix(p, "/")
	for _, dir := range usrMergeDirs {
		if cleaned == dir {
			return true
		}
	}
	return false
}

func normalizeUsrMerge(p string) string {
	for _, dir := range usrMergeDirs {
		if p == dir || strings.HasPrefix(p, dir+"/") {
			return "/usr" + p
		}
	}
	return p
}

func normalizeTarPath(p string) string {
	p = strings.TrimPrefix(p, ".")
	if p != "" && !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func PrefixFilter(prefix, strip string, usrmerge bool, reader *tar.Reader, files []string) error {
	prefix = normalizeTarPath(prefix)

	fileMap := map[string]string{}
	for _, file := range files {
		suffix, ok := strings.CutPrefix(file, strip)
		if !ok {
			return fmt.Errorf("strip prefix %s is not found in %s", strip, file)
		}

		key := suffix
		if idx := strings.Index(key, prefix); idx == -1 {
			key = filepath.Join(prefix, key)
		}

		if !strings.HasPrefix(key, "/") {
			key = "/" + key
		}
		if usrmerge {
			key = normalizeUsrMerge(key)
		}

		log.Tracef("Mapped file %v -> %v", key, file)
		fileMap[key] = file
	}

	if usrmerge {
		prefix = normalizeUsrMerge(prefix)
	}

	for {
		entry, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(fileMap) == 0 {
			break
		}
		name := normalizeTarPath(entry.Name)
		if usrmerge && !(entry.Typeflag == tar.TypeSymlink && isUsrMergeRoot(name)) {
			name = normalizeUsrMerge(name)
		}
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		if _, exists := fileMap[name]; !exists {
			log.Tracef("File %v does not exist in file map", name)
			continue
		}

		log.Tracef("Processing file %v with type %v", fileMap[name], entry.Typeflag)
		switch entry.Typeflag {
		case tar.TypeReg:
			err := func() error {
				writer, err := os.Create(fileMap[name])
				if err != nil {
					return err
				}
				defer writer.Close()
				if _, err := io.Copy(writer, reader); err != nil {
					return err
				}
				return nil
			}()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			linkname := entry.Linkname
			if strings.HasPrefix(linkname, "./") {
				linkname = strings.TrimPrefix(linkname, "./")
			}
			err = os.Symlink(linkname, fileMap[name])
			if err != nil {
				return err
			}
		case tar.TypeLink:
			// In the hard link case, entry.Name will be something like ./usr/include/foo.h, so we
			// just want to get a relative path to the target (entry.Linkname) unlike the symlink case
			// where we just reproduce the entry.Linkname verbatim.
			linkSrc := normalizeTarPath(entry.Name)
			linkTarget := normalizeTarPath(entry.Linkname)
			if usrmerge {
				linkSrc = normalizeUsrMerge(linkSrc)
				linkTarget = normalizeUsrMerge(linkTarget)
			}
			rel, err := filepath.Rel(filepath.Dir(linkSrc), linkTarget)
			if err != nil {
				return err
			}
			log.Tracef("Replacing hard link %v -> %v with relative symlink %v", entry.Name, entry.Linkname, rel)
			if err := os.Symlink(rel, fileMap[name]); err != nil {
				return err
			}
		case tar.TypeDir:
			err = os.MkdirAll(fileMap[name], entry.FileInfo().Mode())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("can't extract %v with type %v: only links, symlinks and files can be specified", fileMap[name], entry.Typeflag)
		}
		delete(fileMap, name)
	}

	if len(fileMap) > 0 {
		return fmt.Errorf("some files could not be found: %v", fileMap)
	}
	return nil
}

func Untar(tmpRoot string, tarFile string) error {
	reader, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	hardLinks := map[string]string{}

	for {
		entry, err := tarReader.Next()
		if err == io.EOF {
			reader.Close()
			break
		} else if err != nil {
			return err
		}

		target := filepath.Join(tmpRoot, entry.Name)
		_, err = filepath.Rel(tmpRoot, target)
		if err != nil {
			return err
		}
		switch entry.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(target, os.ModePerm)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
			err = func() error {
				writer, err := os.Create(target)
				if err != nil {
					return err
				}
				defer writer.Close()
				if _, err := io.Copy(writer, tarReader); err != nil {
					return err
				}
				writer.Close()
				os.Chmod(target, os.FileMode(entry.Mode))
				return nil
			}()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			dir := filepath.Dir(target)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
			linkname := entry.Linkname
			if strings.HasPrefix(linkname, "/") {
				linkname = filepath.Join(tmpRoot, linkname)
				linkname, err = filepath.Rel(filepath.Dir(target), linkname)
			}
			abs := filepath.Join(filepath.Dir(target), linkname)
			if _, err := filepath.Rel(tmpRoot, abs); err != nil {
				return err
			}
			if err = os.Symlink(linkname, target); err != nil {
				return err
			}
		case tar.TypeLink:
			dir := filepath.Dir(target)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
			hardLinks[target] = entry.Linkname
		default:
			log.Debugf("Skipping %s with type %v", entry.Name, entry.Typeflag)
		}
	}

	for target, source := range hardLinks {
		source := filepath.Join(tmpRoot, source)
		if err := os.Link(source, target); err != nil {
			return fmt.Errorf("failed to create hard link from %s to %s: %v", target, source, err)
		}
	}
	return nil
}

func CPIOToTarHeader(entry *cpio.CpioEntry) (*tar.Header, error) {
	tarHeader := &tar.Header{
		Name:     entry.Header.Filename(),
		Size:     entry.Header.Filesize64(),
		Mode:     int64(entry.Header.Mode()),
		Uid:      entry.Header.Uid(),
		Gid:      entry.Header.Gid(),
		ModTime:  time.Unix(int64(entry.Header.Mtime()), 0),
		Devmajor: int64(entry.Header.Devmajor()),
		Devminor: int64(entry.Header.Devminor()),
	}

	switch entry.Header.Mode() &^ 0o7777 {
	case cpio.S_ISCHR:
		tarHeader.Typeflag = tar.TypeChar
	case cpio.S_ISBLK:
		tarHeader.Typeflag = tar.TypeBlock
	case cpio.S_ISDIR:
		tarHeader.Typeflag = tar.TypeDir
	case cpio.S_ISFIFO:
		tarHeader.Typeflag = tar.TypeFifo
	case cpio.S_ISLNK:
		tarHeader.Typeflag = tar.TypeSymlink
		tarHeader.Size = 0
		buf, err := ioutil.ReadAll(entry.Payload)
		if err != nil {
			return nil, err
		}
		tarHeader.Linkname = string(buf)
	case cpio.S_ISREG:
		if entry.Header.Nlink() > 1 && entry.Header.Filesize() == 0 {
			tarHeader.Typeflag = tar.TypeLink
		}
		tarHeader.Typeflag = tar.TypeReg
		tarHeader.Size = 0
	default:
		return nil, fmt.Errorf("unknown file mode 0%o for %s",
			entry.Header.Mode(), entry.Header.Filename())
	}
	return tarHeader, nil
}
