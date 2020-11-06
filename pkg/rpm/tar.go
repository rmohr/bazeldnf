package rpm

import (
	"archive/tar"
	"fmt"
	"io"
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

func PrefixFilter(prefix string, stripPrefix bool, reader *tar.Reader, writer *tar.Writer) error {
	for {
		entry, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if !strings.HasPrefix(entry.Name, prefix) {
			continue
		}
		if stripPrefix {
			relative := strings.HasPrefix(entry.Name, "./")
			entry.Name = strings.TrimPrefix(entry.Name, prefix)
			if relative {
				entry.Name = "." + entry.Name
			}
		}
		if err := writer.WriteHeader(entry); err != nil {
			return err
		}
		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}
	}
}
