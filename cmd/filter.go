package main

import (
	"path/filepath"
	"strings"
)

var allowlist = map[string]struct{}{
	"c":   {},
	"cc":  {},
	"cpp": {},
	"cxx": {},
	"c++": {},
	"C":   {},
	"h":   {},
	"hh":  {},
	"hpp": {},
	"hxx": {},
	"inc": {},
	"inl": {},
	"H":   {},
	"S":   {},
	"a":   {},
	"lo":  {},
	"so":  {},
	"o":   {},
}

var denylist = map[string]struct{}{
	"hmac": {},
}

func filterFiles(files []string) (filteredFiles []string) {
	for _, file := range files {
		s := strings.Split(filepath.Base(file), ".")
		suffix := s[len(s)-1]
		if _, exists := denylist[suffix]; exists {
			continue
		}
		if _, exists := allowlist[suffix]; exists {
			filteredFiles = append(filteredFiles, file)
			continue
		}
		for i := len(s) - 1; i >= 1; i-- {
			if s[i] == "so" {
				filteredFiles = append(filteredFiles, file)
				break
			}
		}

	}
	return
}
