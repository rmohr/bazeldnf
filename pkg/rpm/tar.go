package rpm

import (
	"fmt"
	"io"

	"github.com/sassoftware/go-rpmutils"
	"github.com/sassoftware/go-rpmutils/cpio"
)

func RPMToTar(rpmReader io.Reader, tarWriter io.Writer) (error) {
	rpm, err := rpmutils.ReadRpm(rpmReader)
	if err != nil {
		return fmt.Errorf("failed to read rpm: %s", err)
	}
	payloadReader, err := rpm.PayloadReader()
	if err != nil {
		return fmt.Errorf("failed to open the payload reader: %s", err)
	}
	return cpio.Tar(payloadReader, tarWriter)
}