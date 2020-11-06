package rpm

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestRPMToTar(t *testing.T) {
	tests := []struct {
		name            string
		rpm             string
		expectedHeaders []*tar.Header
		excludedHeaders []*tar.Header
		wantErr         bool
		prefix          string
		stripPrefix bool
	}{
		{
			name:    "should convert a RPM to tar and keep all entries",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "libvirt-libs-6.1.0-2.fc32.x86_64.rpm/rpm/downloaded"),
			wantErr: false,
			expectedHeaders: []*tar.Header{
				{Name: "./etc/libvirt/libvirt-admin.conf", Size: 450, Mode: 33188},
				{Name: "./etc/libvirt/libvirt.conf", Size: 547, Mode: 33188},
				{Name: "./usr/lib64/libvirt.so.0", Size: 19, Mode: 41471},
			},
			prefix: "",
			stripPrefix: false,
		},
		{
			name:    "should convert a RPM to tar and only keep shared librarires",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "libvirt-libs-6.1.0-2.fc32.x86_64.rpm/rpm/downloaded"),
			wantErr: false,
			expectedHeaders: []*tar.Header{
				{Name: "./usr/lib64/libvirt.so.0", Size: 19, Mode: 41471},
			},
			excludedHeaders: []*tar.Header{
				{Name: "./etc/libvirt/libvirt-admin.conf", Size: 450, Mode: 33188},
				{Name: "./etc/libvirt/libvirt.conf", Size: 547, Mode: 33188},
			},
			prefix: "./usr/lib64",
			stripPrefix: false,
		},
		{
			name:    "should convert a RPM to tar and only keep shared librarires and strip their prefix",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "libvirt-libs-6.1.0-2.fc32.x86_64.rpm/rpm/downloaded"),
			wantErr: false,
			expectedHeaders: []*tar.Header{
				{Name: "./libvirt.so.0", Size: 19, Mode: 41471},
			},
			excludedHeaders: []*tar.Header{
				{Name: "./etc/libvirt/libvirt-admin.conf", Size: 450, Mode: 33188},
				{Name: "./etc/libvirt/libvirt.conf", Size: 547, Mode: 33188},
				{Name: "./libvirt-admin.conf", Size: 450, Mode: 33188},
				{Name: "./libvirt.conf", Size: 547, Mode: 33188},
				{Name: "./usr/lib64/libvirt.so.0", Size: 19, Mode: 41471},
			},
			prefix: "./usr/lib64",
			stripPrefix: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			f, err := os.Open(tt.rpm)
			g.Expect(err).ToNot(HaveOccurred())
			defer f.Close()

			tmpdir, err := ioutil.TempDir("", "tar")
			g.Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir)
			tarWriter, err := os.Create(filepath.Join(tmpdir, "test.tar"))
			g.Expect(err).ToNot(HaveOccurred())

			pipeReader, pipeWriter := io.Pipe()
			defer pipeReader.Close()
			defer pipeWriter.Close()

			defer tarWriter.Close()
			go func() {
				err := RPMToTar(f, tar.NewWriter(pipeWriter))
				g.Expect(err).ToNot(HaveOccurred())
				pipeWriter.Close()
			}()

			err = PrefixFilter(tt.prefix, tt.stripPrefix, tar.NewReader(pipeReader), tar.NewWriter(tarWriter))
			g.Expect(err).ToNot(HaveOccurred())

			tarReader, err := os.Open(filepath.Join(tmpdir, "test.tar"))
			g.Expect(err).ToNot(HaveOccurred())
			tarBall := tar.NewReader(tarReader)
			discoveredHeaders := []*tar.Header{}
			for {
				header, err := tarBall.Next()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						g.Expect(err).ToNot(HaveOccurred())
					}
				}
				discoveredHeaders = append(discoveredHeaders, &tar.Header{
					Name: header.Name,
					Size: header.Size,
					Uid: header.Uid,
					Gid: header.Gid,
					Mode: header.Mode,
				})
			}
			g.Expect(discoveredHeaders).To(ContainElements(tt.expectedHeaders))
			if len(tt.excludedHeaders) > 0 {
				g.Expect(discoveredHeaders).ToNot(ContainElements(tt.excludedHeaders))
			}
		})
	}
}
