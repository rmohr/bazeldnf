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
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "libvirt-libs-11.0.0-1.fc42.x86_64.rpm/rpm/downloaded"),
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
		name     string
		rpm      string
		files    []string
		expected []fileInfo
		wantErr  bool
		prefix   string
	}{
		{
			name:    "should extract a symlink from a tar archive",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "libvirt-libs-11.0.0-1.fc42.x86_64.rpm/rpm/downloaded"),
			wantErr: false,
			files:   []string{"/usr/lib64/libvirt.so.0"},
			expected: []fileInfo{
				{Name: "usr", Size: 96, Children: []fileInfo{
					{Name: "lib64", Size: 96, Children: []fileInfo{
						{Name: "libvirt.so.0", Size: 20},
					}},
				}},
			},
			prefix: "./usr/lib64",
		},
		{
			name:    "should extract multiple files from a tar archive",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "libvirt-libs-11.0.0-1.fc42.x86_64.rpm/rpm/downloaded"),
			wantErr: false,
			files: []string{
				"/etc/libvirt/libvirt-admin.conf",
				"/etc/libvirt/libvirt.conf",
			},
			expected: []fileInfo{
				{Name: "etc", Size: 96, Children: []fileInfo{
					{Name: "libvirt", Size: 128, Children: []fileInfo{
						{Name: "libvirt-admin.conf", Size: 450},
						{Name: "libvirt.conf", Size: 547},
					}},
				}},
			},
			prefix: "./etc/libvirt/",
		},
		{
			name:    "should extract multiple files with the same name from a tar archive",
			rpm:     filepath.Join(os.Getenv("TEST_SRCDIR"), "abseil-cpp-devel-20240722.1-1.fc42.x86_64.rpm/rpm/downloaded"),
			wantErr: false,
			files: []string{
				"/usr/include/absl/log/globals.h",
				"/usr/include/absl/log/internal/globals.h",
			},
			expected: []fileInfo{
				{Name: "usr", Size: 96, Children: []fileInfo{
					{Name: "include", Size: 96, Children: []fileInfo{
						{Name: "absl", Size: 96, Children: []fileInfo{
							{Name: "log", Size: 128, Children: []fileInfo{
								{Name: "globals.h", Size: 8391},
								{Name: "internal", Size: 96, Children: []fileInfo{
									{Name: "globals.h", Size: 4030},
								}},
							}},
						}},
					}},
				}},
			},
			prefix: "./usr/include/",
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
			for _, file := range tt.files {
				// bazel rules automagically create these directories
				err = os.MkdirAll(filepath.Join(tmpdir, filepath.Dir(file)), 0777)
				g.Expect(err).ToNot(HaveOccurred())
				files = append(files, filepath.Join(tmpdir, file))
			}

			err = PrefixFilter(tt.prefix, tar.NewReader(pipeReader), files)
			g.Expect(err).ToNot(HaveOccurred())

			discoveredHeaders, err := collectFileInfo(tmpdir)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(discoveredHeaders).To(ConsistOf(tt.expected))
		})
	}
}

func collectFileInfo(dirName string) ([]fileInfo, error) {
	r := []fileInfo{}

	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		return r, err
	}

	for _, file := range fileInfos {
		fileInfo := fileInfo{
			Name: file.Name(),
			Size: file.Size(),
		}

		if file.IsDir() {
			children, err := collectFileInfo(filepath.Join(dirName, file.Name()))
			if err != nil {
				return r, err
			}
			fileInfo.Children = children
		}

		r = append(r, fileInfo)
	}

	return r, nil
}

type fileInfo struct {
	Name     string
	Size     int64
	Children []fileInfo
}
