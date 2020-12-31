package rpm

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sassoftware/go-rpmutils"
	"github.com/sassoftware/go-rpmutils/cpio"
	log "github.com/sirupsen/logrus"
)

func RPMToTar(rpmReader io.Reader, tarWriter *tar.Writer, noSymlinksAndDirs bool, capabilities map[string][]string) error {
	rpm, err := rpmutils.ReadRpm(rpmReader)
	if err != nil {
		return fmt.Errorf("failed to read rpm: %s", err)
	}
	payloadReader, err := rpm.RawUncompressedRPMPayloadReader()
	if err != nil {
		return fmt.Errorf("failed to open the payload reader: %s", err)
	}
	return Tar(payloadReader, tarWriter, noSymlinksAndDirs, capabilities)
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

func RPMReader(rpmReader io.Reader, tarWriter *tar.Writer) error {
	rpm, err := rpmutils.ReadRpm(rpmReader)
	if err != nil {
		return fmt.Errorf("failed to read rpm: %s", err)
	}
	payloadReader, err := rpm.RawUncompressedRPMPayloadReader()
	if err != nil {
		return fmt.Errorf("failed to open the payload reader: %s", err)
	}
	return cpio.Tar(payloadReader, tarWriter)
}

func PrefixFilter(prefix string, reader *tar.Reader, files []string) error {
	prefix = strings.TrimPrefix(prefix, ".")

	fileMap := map[string]string{}
	for _, file := range files {
		fileMap[filepath.Base(file)] = file
	}
	for {
		entry, err := reader.Next()
		if err == io.EOF {
			break
		}
		if len(fileMap) == 0 {
			break
		}
		name := strings.TrimPrefix(entry.Name, ".")
		if strings.HasPrefix(name, prefix) {
		} else if prefix == "/usr/lib64" && strings.HasPrefix(name, "/lib64") {
		} else {
			continue
		}
		basename := filepath.Base(name)
		if _, exists := fileMap[basename]; !exists {
			continue
		}
		if entry.Typeflag == tar.TypeReg {
			err := func() error {
				writer, err := os.Create(fileMap[basename])
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
			delete(fileMap, basename)
		} else if entry.Typeflag == tar.TypeSymlink {
			linkname := strings.TrimPrefix(entry.Linkname, ".")
			err = os.Symlink(linkname, fileMap[basename])
			if err != nil {
				return err
			}
			delete(fileMap, basename)
		} else {
			return fmt.Errorf("can't extract %s, only symlinks and files can be specified", fileMap[basename])
		}
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
				if _, err := io.Copy(writer, tarReader); err != nil {
					return err
				}
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
			linkname := strings.TrimPrefix(entry.Linkname, ".")
			err = os.Symlink(linkname, target)
			if err != nil {
				return err
			}
		default:
			log.Debugf("Skipping %s with type %v", entry.Name, entry.Typeflag)
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

	switch entry.Header.Mode() &^ 07777 {
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
		buf := make([]byte, entry.Header.Filesize())
		if _, err := entry.Payload.Read(buf); err != nil {
			return nil, err
		}
		tarHeader.Linkname = string(buf)
	case cpio.S_ISREG:
		if entry.Header.Nlink() > 1 && entry.Header.Filesize() == 0 {
			tarHeader.Typeflag = tar.TypeLink
		}
		tarHeader.Typeflag = tar.TypeReg
	default:
		return nil, fmt.Errorf("unknown file mode 0%o for %s",
			entry.Header.Mode(), entry.Header.Filename())
	}
	return tarHeader, nil
}
