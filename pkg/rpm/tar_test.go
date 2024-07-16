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
		stripPrefix     bool
	}{
		{
			name:    "should convert a RPM to tar and keep all entries",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_RPM")),
			wantErr: false,
			expectedHeaders: []*tar.Header{
				{Name: "./etc/libvirt/libvirt-admin.conf", Size: 450, Mode: 33188},
				{Name: "./etc/libvirt/libvirt.conf", Size: 547, Mode: 33188},
				{Name: "./usr/lib64/libvirt.so.0", Size: 0, Mode: 41471},
			},
			prefix:      "",
			stripPrefix: false,
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
			defer tarWriter.Close()

			collector := NewCollector()
			err = collector.RPMToTar(f, tar.NewWriter(tarWriter), false, nil, nil)
			g.Expect(err).ToNot(HaveOccurred())
			tarWriter.Close()

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
					Uid:  header.Uid,
					Gid:  header.Gid,
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

func TestTar2Files(t *testing.T) {
	tests := []struct {
		name          string
		rpm           string
		expectedFiles []*fileInfo
		wantErr       bool
		prefix        string
	}{
		{
			name:    "should extract a symlink from a tar archive",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_RPM")),
			wantErr: false,
			expectedFiles: []*fileInfo{
				{Name: "libvirt.so.0", Size: 19},
			},
			prefix: "./usr/lib64",
		},
		{
			name:    "should extract multiple files from a tar archive",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_RPM")),
			wantErr: false,
			expectedFiles: []*fileInfo{
				{Name: "libvirt-admin.conf", Size: 450},
				{Name: "libvirt.conf", Size: 547},
			},
			prefix: "./etc/libvirt/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			f, err := os.Open(tt.rpm)
			g.Expect(err).ToNot(HaveOccurred())
			defer f.Close()

			tmpdir, err := ioutil.TempDir("", "files")
			g.Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir)

			pipeReader, pipeWriter := io.Pipe()
			defer pipeReader.Close()
			defer pipeWriter.Close()

			collector := NewCollector()
			go func() {
				_ = collector.RPMToTar(f, tar.NewWriter(pipeWriter), false, nil, nil)
				pipeWriter.Close()
			}()

			files := []string{}
			for _, file := range tt.expectedFiles {
				files = append(files, filepath.Join(tmpdir, file.Name))
			}

			err = PrefixFilter(tt.prefix, tar.NewReader(pipeReader), files)
			g.Expect(err).ToNot(HaveOccurred())

			discoveredHeaders := []*fileInfo{}
			fileInfos, err := ioutil.ReadDir(tmpdir)
			g.Expect(err).ToNot(HaveOccurred())
			for _, file := range fileInfos {
				discoveredHeaders = append(discoveredHeaders, &fileInfo{
					Name: file.Name(),
					Size: file.Size(),
				})
			}
			g.Expect(discoveredHeaders).To(ConsistOf(tt.expectedFiles))
		})
	}
}

type fileInfo struct {
	Name string
	Size int64
}
