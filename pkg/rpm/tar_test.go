package rpm

import (
	"archive/tar"
	"bytes"
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
		prefix   string
	}{
		{
			name:  "should extract a symlink from a tar archive",
			rpm:   libvirtLibsRpm,
			files: []string{"/usr/lib64/libvirt.so.0"},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "lib64", Children: []fileInfo{
						{Name: "libvirt.so.0", Link: "libvirt.so.0.11000.0", Size: 20},
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
						{Name: "com_err.h", Size: 12, Link: "et/com_err.h"},
						{Name: "et", Children: []fileInfo{
							{Name: "com_err.h", Size: 2118},
						}},
					}},
				}},
			},
			prefix: "./usr/include/",
		},
		{
			name:  "missing link target from a tar archive",
			rpm:   libComErrDevelRpm,
			files: []string{"/usr/include/com_err.h"},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "include", Children: []fileInfo{
						{Name: "com_err.h", Size: 12, Link: "et/com_err.h"},
					}},
				}},
			},
			prefix: "./usr/include/",
		},
		{
			name:  "extract a dir from a tar archive",
			rpm:   libvirtLibsRpm,
			files: []string{"/usr/share/libvirt"},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "share", Children: []fileInfo{
						{Name: "libvirt", Children: []fileInfo{}},
					}},
				}},
			},
			prefix: "./usr/share/",
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

			err = PrefixFilter(tt.prefix, tmpdir, true, tar.NewReader(pipeReader), files)
			g.Expect(err).ToNot(HaveOccurred())
			discoveredHeaders, err := collectFileInfo(tmpdir)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(discoveredHeaders).To(ConsistOf(tt.expected))
		})
	}
}

type tarTestEntry struct {
	name     string
	typeflag byte
	body     string
	linkname string
}

func buildTestTar(t *testing.T, entries []tarTestEntry) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeflag,
			Mode:     0o755,
			Size:     int64(len(e.body)),
			Linkname: e.linkname,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tw.WriteHeader(%q): %v", e.name, err)
		}
		if len(e.body) > 0 {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatalf("tw.Write(%q): %v", e.name, err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tw.Close(): %v", err)
	}
	return &buf
}

func TestPrefixFilterUsrMerge(t *testing.T) {
	const payload = "elf-binary-payload"
	payloadLen := int64(len(payload))

	tests := []struct {
		name     string
		prefix   string
		usrmerge bool
		entries  []tarTestEntry
		files    []string
		expected []fileInfo
		wantErr  bool
	}{
		{
			name:     "lib64_in_tar_extracted_under_usr_lib64_before_root_symlink",
			prefix:   "/usr/lib64",
			usrmerge: true,
			entries: []tarTestEntry{
				{name: "./lib64/libtirpc.so.3", typeflag: tar.TypeSymlink, linkname: "libtirpc.so.3.0.0"},
				{name: "./lib64/libtirpc.so.3.0.0", typeflag: tar.TypeReg, body: payload},
				{name: "./lib64", typeflag: tar.TypeSymlink, linkname: "usr/lib64"},
			},
			files: []string{
				"/usr/lib64/libtirpc.so.3",
				"/usr/lib64/libtirpc.so.3.0.0",
			},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "lib64", Children: []fileInfo{
						{Name: "libtirpc.so.3", Link: "libtirpc.so.3.0.0", Size: int64(len("libtirpc.so.3.0.0"))},
						{Name: "libtirpc.so.3.0.0", Size: payloadLen},
					}},
				}},
			},
		},
		{
			name:     "lib64_in_tar_extracted_under_usr_lib64_after_root_symlink",
			prefix:   "/usr/lib64",
			usrmerge: true,
			entries: []tarTestEntry{
				{name: "./lib64", typeflag: tar.TypeSymlink, linkname: "usr/lib64"},
				{name: "./lib64/libtirpc.so.3.0.0", typeflag: tar.TypeReg, body: payload},
			},
			files: []string{"/usr/lib64/libtirpc.so.3.0.0"},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "lib64", Children: []fileInfo{
						{Name: "libtirpc.so.3.0.0", Size: payloadLen},
					}},
				}},
			},
		},
		{
			name:     "hardlink_across_bin_and_usr_bin",
			prefix:   "/usr/bin",
			usrmerge: true,
			entries: []tarTestEntry{
				{name: "./usr/bin/gawk", typeflag: tar.TypeReg, body: payload},
				{name: "./bin/awk", typeflag: tar.TypeLink, linkname: "./usr/bin/gawk"},
			},
			files: []string{
				"/usr/bin/awk",
				"/usr/bin/gawk",
			},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "bin", Children: []fileInfo{
						{Name: "awk", Link: "gawk", Size: int64(len("gawk"))},
						{Name: "gawk", Size: payloadLen},
					}},
				}},
			},
		},
		{
			name:     "usrmerge_disabled_does_not_match_lib64_for_usr_lib64",
			prefix:   "/usr/lib64",
			usrmerge: false,
			entries: []tarTestEntry{
				{name: "./lib64/libtirpc.so.3.0.0", typeflag: tar.TypeReg, body: payload},
			},
			files:   []string{"/usr/lib64/libtirpc.so.3.0.0"},
			wantErr: true,
		},
		{
			name:     "entries_and_prefix_without_leading_dot_slash",
			prefix:   "usr/bin",
			usrmerge: true,
			entries: []tarTestEntry{
				{name: "usr/bin/gawk", typeflag: tar.TypeReg, body: payload},
				{name: "bin/awk", typeflag: tar.TypeLink, linkname: "usr/bin/gawk"},
			},
			files: []string{
				"/usr/bin/awk",
				"/usr/bin/gawk",
			},
			expected: []fileInfo{
				{Name: "usr", Children: []fileInfo{
					{Name: "bin", Children: []fileInfo{
						{Name: "awk", Link: "gawk", Size: int64(len("gawk"))},
						{Name: "gawk", Size: payloadLen},
					}},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			tmpdir := t.TempDir()

			var outFiles []string
			for _, f := range tt.files {
				fullPath := filepath.Join(tmpdir, f)
				g.Expect(os.MkdirAll(filepath.Dir(fullPath), 0o755)).To(Succeed())
				outFiles = append(outFiles, fullPath)
			}

			tarBuf := buildTestTar(t, tt.entries)
			err := PrefixFilter(tt.prefix, tmpdir, tt.usrmerge, tar.NewReader(tarBuf), outFiles)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			got, err := collectFileInfo(tmpdir)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(got).To(ConsistOf(tt.expected))
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
			if file.Mode()&os.ModeSymlink != 0 {
				symlink := filepath.Join(dirName, file.Name())
				if target, err := os.Readlink(symlink); err == nil {
					fileInfo.Link = target
				}
			}
		}

		r = append(r, fileInfo)
	}

	return r, nil
}

type fileInfo struct {
	Name     string
	Size     int64
	Link     string
	Children []fileInfo
}
