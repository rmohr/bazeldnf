/*
 * Copyright (c) SAS Institute, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpm

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"github.com/rmohr/bazeldnf/pkg/xattr"
	"github.com/sassoftware/go-rpmutils/cpio"
	"github.com/sirupsen/logrus"
)

// Extract the contents of a cpio stream from and writes it as a tar file into the provided writer
func Tar(rs io.Reader, tarfile *tar.Writer, noSymlinksAndDirs bool, capabilities map[string][]string, selinuxLabels map[string]string, createdPaths map[string]struct{}) error {
	hardLinks := map[int][]*tar.Header{}
	inodes := map[int]string{}

	stream := cpio.NewCpioStream(rs)

	for {
		entry, err := stream.ReadNextEntry()
		if err != nil {
			return err
		}

		if entry.Header.Filename() == cpio.TRAILER {
			break
		}

		if entry.Header.Filename() != "" {
			if _, exists := createdPaths[entry.Header.Filename()]; exists {
				logrus.Debugf("Skipping duplicate tar entry %s\n", entry.Header.Filename())
				continue
			}
			createdPaths[entry.Header.Filename()] = struct{}{}
		}

		pax := map[string]string{}
		if caps, exists := capabilities[entry.Header.Filename()]; exists {
			if err := xattr.AddCapabilities(pax, caps); err != nil {
				return fmt.Errorf("failed setting capabilities on %s: %v", entry.Header.Filename(), err)
			}
		}
		if label, exists := selinuxLabels[entry.Header.Filename()]; exists {
			if err := xattr.SetSELinuxLabel(pax, label); err != nil {
				return fmt.Errorf("failed setting selinux label on %s: %v", entry.Header.Filename(), err)
			}
		}

		tarHeader := &tar.Header{
			Name:       entry.Header.Filename(),
			Size:       entry.Header.Filesize64(),
			Mode:       int64(entry.Header.Mode()),
			Uid:        entry.Header.Uid(),
			Gid:        entry.Header.Gid(),
			ModTime:    time.Unix(int64(entry.Header.Mtime()), 0),
			Devmajor:   int64(entry.Header.Devmajor()),
			Devminor:   int64(entry.Header.Devminor()),
			PAXRecords: pax,
		}

		var payload io.Reader
		switch entry.Header.Mode() &^ 0o7777 {
		case cpio.S_ISCHR:
			tarHeader.Typeflag = tar.TypeChar
		case cpio.S_ISBLK:
			tarHeader.Typeflag = tar.TypeBlock
		case cpio.S_ISDIR:
			if noSymlinksAndDirs {
				continue
			}
			tarHeader.Typeflag = tar.TypeDir
		case cpio.S_ISFIFO:
			tarHeader.Typeflag = tar.TypeFifo
		case cpio.S_ISLNK:
			if noSymlinksAndDirs {
				continue
			}
			tarHeader.Typeflag = tar.TypeSymlink
			tarHeader.Size = 0
			buf, err := ioutil.ReadAll(entry.Payload)
			if err != nil {
				return err
			}
			tarHeader.Linkname = string(buf)
		case cpio.S_ISREG:
			if entry.Header.Nlink() > 1 && entry.Header.Filesize() == 0 {
				tarHeader.Typeflag = tar.TypeLink
				tarHeader.Size = 0
				hardLinks[entry.Header.Ino()] = append(hardLinks[entry.Header.Ino()], tarHeader)
				continue
			}
			tarHeader.Typeflag = tar.TypeReg
			payload = entry.Payload
			inodes[entry.Header.Ino()] = entry.Header.Filename()
		default:
			return fmt.Errorf("unknown file mode 0%o for %s",
				entry.Header.Mode(), entry.Header.Filename())
		}
		if err := tarfile.WriteHeader(tarHeader); err != nil {
			return fmt.Errorf("could not write tar header for %v: %v", tarHeader.Name, err)
		}
		if payload != nil {
			written, err := io.Copy(tarfile, entry.Payload)
			if err != nil {
				return fmt.Errorf("could not write body for %v: %v", tarHeader.Name, err)
			}
			if written != int64(entry.Header.Filesize()) {
				return fmt.Errorf("short write body for %v", tarHeader.Name)
			}
		}
	}
	// write hardlinks
	var sortedNodes []int
	for node := range hardLinks {
		sortedNodes = append(sortedNodes, node)
	}
	sort.Ints(sortedNodes)
	for _, node := range sortedNodes {
		links := hardLinks[node]
		target := inodes[node]
		if target == "" {
			return fmt.Errorf("no target file for inode %v found", node)
		}
		for _, tarHeader := range links {
			tarHeader.Linkname = target
			if err := tarfile.WriteHeader(tarHeader); err != nil {
				return fmt.Errorf("could not write tar header for %v", tarHeader.Name)
			}
		}
	}

	return nil
}
