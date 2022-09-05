package ldd

import (
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
)

func Resolve(objects []string, library_path []string) (finalFiles []string, err error) {
	discovered := map[string]struct{}{}
	for _, obj := range objects {
		if files, err := resolve(obj, library_path); err != nil {
			return nil, err
		} else {
			for _, l := range files {
				if _, exists := discovered[l]; !exists {
					discovered[l] = struct{}{}
					finalFiles = append(finalFiles, l)
				}
			}
		}
	}
	return
}

func resolve(library string, library_path []string) ([]string, error) {

	next := []string{library}
	processed := map[string]struct{}{}

	finalFiles := []string{}
	for {
		if len(next) == 0 {
			break
		}
		if d, err := ldd(next[0], library_path); err != nil {
			return nil, err
		} else {
			for _, l := range d {
				if _, exists := processed[l]; !exists {
					next = append(next, l)
					processed[l] = struct{}{}
					finalFiles = append(finalFiles, l)
				}
			}
			if len(next) > 1 {
				next = next[1 : len(next)-1]
			} else {
				next = []string{}
			}
		}
	}

	finalFiles = append(finalFiles, library)
	symlinks, err := followSymlinks(library)
	if err != nil {
		return nil, err
	}
	finalFiles = append(finalFiles, symlinks...)

	return finalFiles, nil
}

func ldd(library string, library_path []string) (discovered []string, err error) {
	bin, err := elf.Open(library)
	if err != nil {
		return nil, err
	}
	libs, err := bin.ImportedLibraries()
	if err != nil {
		return nil, err
	}
	for _, l := range libs {
		found := false
		for _, dir := range library_path {
			_, err := os.Stat(filepath.Join(dir, l))
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			discovered = append(discovered, filepath.Join(dir, l))
			symlinks, err := followSymlinks(filepath.Join(dir, l))
			if err != nil {
				return nil, err
			}
			discovered = append(discovered, symlinks...)
			found = true
			break
		}
		if !found {
			return nil, fmt.Errorf("%v not found in any of %v", l, library_path)
		}
	}
	return discovered, nil
}

func followSymlinks(file string) (files []string, err error) {
	dir := filepath.Dir(file)
	for {
		info, err := os.Lstat(file)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			file, err = os.Readlink(file)
			if err != nil {
				return nil, err
			}
			if !filepath.IsAbs(file) {
				file = filepath.Join(dir, file)
			}
			files = append(files, file)
		} else {
			files = append(files, file)
			break
		}
	}
	return
}
