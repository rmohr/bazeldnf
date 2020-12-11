package rpm

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sassoftware/go-rpmutils"
	"github.com/sassoftware/go-rpmutils/cpio"
)

func RPMToTar(rpmReader io.Reader, tarWriter *tar.Writer) error {
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
		if !strings.HasPrefix(name, prefix) {
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

	for _, t := range []byte{tar.TypeDir, tar.TypeSymlink, tar.TypeReg} {

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

			if strings.HasSuffix(entry.Name, "lib64") {
				fmt.Println(entry.Name)
			}
			if entry.Typeflag != t {
				continue
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
				err := func() error {
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
				linkname := strings.TrimPrefix(entry.Linkname, ".")
				err = os.Symlink(linkname, target)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("file type %v not supported right now", entry.Typeflag)
			}
		}
	}
	return nil
}
