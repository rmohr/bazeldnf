package rpm

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	. "github.com/onsi/gomega"
)

func TestRPMToTar(t *testing.T) {
	libvirtLibsRpm, err := runfiles.Rlocation(os.Getenv("LIBVIRT_LIBS"))
	if err != nil {
		panic(err)
	}

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
			rpm:     libvirtLibsRpm,
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
	abseilCppDevelRpm, err := runfiles.Rlocation(os.Getenv("CPP_DEVEL"))
	if err != nil {
		t.Fatal(err)
	}

	libComErrDevelRpm, err := runfiles.Rlocation(os.Getenv("LIBCOM_ERR_DEVEL"))
	if err != nil {
		t.Fatal(err)
	}

	libvirtLibsRpm, err := runfiles.Rlocation(os.Getenv("LIBVIRT_LIBS"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		rpm      string
		files    []string
		expected []fileInfo
		wantErr  string
		prefix   string
	}{
		{
			name:  "should extract a symlink from a tar archive",
			rpm:   libvirtLibsRpm,
			files: []string{"/usr/lib64/libvirt.so.0"},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "lib64", Children: []fileInfo{
						{Name: "libvirt.so.0", Size: 20},
					}},
				}},
			},
			prefix: "./usr/lib64",
		},
		{
			name: "should extract multiple files from a tar archive",
			rpm:  libvirtLibsRpm,
			files: []string{
				"/etc/libvirt/libvirt-admin.conf",
				"/etc/libvirt/libvirt.conf",
			},
			expected: []fileInfo{
				{Name: "etc", Children: []fileInfo{
					{Name: "libvirt", Children: []fileInfo{
						{Name: "libvirt-admin.conf", Size: 450},
						{Name: "libvirt.conf", Size: 547},
					}},
				}},
			},
			prefix: "./etc/libvirt/",
		},
		{
			name: "should extract multiple files with the same name from a tar archive",
			rpm:  abseilCppDevelRpm,
			files: []string{
				"/usr/include/absl/log/globals.h",
				"/usr/include/absl/log/internal/globals.h",
			},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "include", Children: []fileInfo{
						{Name: "absl", Children: []fileInfo{
							{Name: "log", Children: []fileInfo{
								{Name: "globals.h", Size: 8391},
								{Name: "internal", Children: []fileInfo{
									{Name: "globals.h", Size: 4030},
								}},
							}},
						}},
					}},
				}},
			},
			prefix: "./usr/include/",
		},
		{
			name: "should extract link from a tar archive",
			rpm:  libComErrDevelRpm,
			files: []string{
				"/usr/include/com_err.h",
				"/usr/include/et/com_err.h",
			},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "include", Children: []fileInfo{
						{Name: "com_err.h", Size: 2118},
						{Name: "et", Children: []fileInfo{
							{Name: "com_err.h", Size: 2118},
						}},
					}},
				}},
			},
			prefix: "./usr/include/",
		},
		{
			name:    "missing link target from a tar archive",
			rpm:     libComErrDevelRpm,
			files:   []string{"/usr/include/com_err.h"},
			wantErr: "/usr/include/com_err.h is a link, but the link target /usr/include/et/com_err.h is not included in the filtered output",
			prefix:  "./usr/include/",
		}}
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

			err = PrefixFilter(tt.prefix, tmpdir, tar.NewReader(pipeReader), files)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
				return
			}
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
		}

		if file.IsDir() {
			children, err := collectFileInfo(filepath.Join(dirName, file.Name()))
			if err != nil {
				return r, err
			}
			fileInfo.Children = children
		} else {
			fileInfo.Size = file.Size()
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
