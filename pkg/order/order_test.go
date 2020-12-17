package order

import (
	"archive/tar"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNode_Traverse(t *testing.T) {
	tests := []struct {
		name         string
		givenHeaders [][]tar.Header
		wantHeaders  []tar.Header
	}{
		{
			name: "should sort directories by breadth first",
			wantHeaders: []tar.Header{
				{
					Name:     "/usr/lib",
					Typeflag: tar.TypeSymlink,
				},
				{
					Name:     "/usr/lib64",
					Typeflag: tar.TypeSymlink,
				},
				{
					Name:     "/usr/a",
					Typeflag: tar.TypeDir,
				},
				{
					Name:     "/usr/b",
					Typeflag: tar.TypeDir,
				},
				{
					Name:     "/usr/a/b",
					Typeflag: tar.TypeSymlink,
				},
				{
					Name:     "/usr/lib/a/b/c/",
					Typeflag: tar.TypeSymlink,
				},
				{
					Name:     "/usr/lib/a/b/c/d/e/",
					Typeflag: tar.TypeDir,
				},
			},
			givenHeaders: [][]tar.Header{
				{
					tar.Header{
						Name:     "/usr/lib/a/b/c/d/e/",
						Typeflag: tar.TypeDir,
					},
					tar.Header{
						Name:     "/usr/lib",
						Typeflag: tar.TypeSymlink,
					},
					tar.Header{
						Name:     "/usr/lib64",
						Typeflag: tar.TypeDir,
					},
					tar.Header{
						Name:     "/usr/lib/a/b/c/",
						Typeflag: tar.TypeSymlink,
					},
					tar.Header{
						Name:     "/usr/a/b",
						Typeflag: tar.TypeSymlink,
					},
				},
				{
					{
						Name:     "/usr/a",
						Typeflag: tar.TypeDir,
					},
					{
						Name:     "/usr/b",
						Typeflag: tar.TypeDir,
					},
					tar.Header{
						Name:     "/usr/lib64",
						Typeflag: tar.TypeSymlink,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			n := NewDirectoryTree()
			for _, headers := range tt.givenHeaders {
				n.Add(headers)
			}
			g.Expect(n.Traverse()).To(Equal(tt.wantHeaders))
		})
	}
}
