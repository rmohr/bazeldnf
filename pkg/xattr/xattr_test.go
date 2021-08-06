package xattr

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

var g *GomegaWithT

func TestSettingSELinuxLabel(t *testing.T) {
	g = NewGomegaWithT(t)
	referenceEntry, err := getHeader("blub")
	g.Expect(err).ToNot(HaveOccurred())

	generatedEntry := &tar.Header{Name: "blub"}
	labels := map[string]string{"blub": "unconfined_u:object_r:user_home_t:s0", "somethingelse": "something"}

	g.Expect(enrichEntry(generatedEntry, nil, labels)).To(Succeed())

	g.Expect(generatedEntry.PAXRecords[selinux_header]).To(Equal(referenceEntry.PAXRecords[selinux_header]))
}

func getHeader(name string) (*tar.Header, error) {
	f, err := os.Open("testdata/xattr.tar")
	g.Expect(err).ToNot(HaveOccurred())
	defer f.Close()
	r := tar.NewReader(f)
	for {
		entry, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			g.Expect(err).ToNot(HaveOccurred())
		}
		if entry.Name == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("entry %v does not exist", name)
}
