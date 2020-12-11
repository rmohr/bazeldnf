package order

import (
	"archive/tar"
	"path/filepath"
	"strings"
)

type Node struct {
	name     string
	Header   *tar.Header
	siblings map[string]*Node
	keys     []string
}

func (n *Node) Add(headers []tar.Header) {
	for j, h := range headers {
		if h.Typeflag != tar.TypeDir && h.Typeflag != tar.TypeSymlink {
			continue
		}
		paths := strings.Split(strings.TrimPrefix(filepath.Clean(h.Name), "/"), "/")
		node := n
		for i, p := range paths {
			if _, exists := node.siblings[p]; !exists {
				node.siblings[p] = &Node{
					name:     p,
					Header:   nil,
					siblings: map[string]*Node{},
				}
				node.keys = append(node.keys, p)
			}
			if len(paths)-1 == i {
				if node.siblings[p].Header == nil || node.siblings[p].Header.Typeflag == tar.TypeDir {
					node.siblings[p].Header = &headers[j]
				}
			}
			node = node.siblings[p]
		}
	}
}

func (n *Node) Traverse() (headers []tar.Header) {
	var queue []*Node
	for _, k := range n.keys {
		queue = append(queue, n.siblings[k])
	}

	for {
		if len(queue) == 0 {
			break
		}
		next := queue[0]
		queue = queue[1:]
		if next.Header != nil {
			headers = append(headers, *next.Header)
		}
		for _, k := range next.keys {
			queue = append(queue, next.siblings[k])
		}
	}
	return
}

func NewDirectoryTree() *Node {
	return &Node{
		siblings: map[string]*Node{},
	}
}
