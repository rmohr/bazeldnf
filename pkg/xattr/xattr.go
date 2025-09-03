package xattr

import (
	"archive/tar"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const (
	capabilities_header = "SCHILY.xattr.security.capability"
	selinux_header      = "SCHILY.xattr.security.selinux"
)

var cap_empty_bitmask = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var supported_capabilities = map[string][]byte{
	"cap_chown":            {1, 0, 0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"cap_net_bind_service": {1, 0, 0, 2, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"cap_net_admin":        {1, 0, 0, 2, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"cap_net_raw":          {1, 0, 0, 2, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"cap_sys_ptrace":       {1, 0, 0, 2, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

func AddCapabilities(pax map[string]string, capabilities []string) error {
	for _, cap := range capabilities {
		if _, supported := supported_capabilities[cap]; !supported {
			return fmt.Errorf("requested capability '%s' is not supported", cap)
		}
		if _, exists := pax[capabilities_header]; !exists {
			pax[capabilities_header] = string(cap_empty_bitmask)
		}
		val := []byte(pax[capabilities_header])
		for i, b := range supported_capabilities[cap] {
			val[i] = val[i] | b
		}
		pax[capabilities_header] = string(val)
	}
	return nil
}

func SetSELinuxLabel(pax map[string]string, label string) error {
	if label == "" {
		return fmt.Errorf("label must not be empty, but got '%s'", label)
	}
	pax[selinux_header] = fmt.Sprintf("%s\x00", label)
	return nil
}

func Apply(reader *tar.Reader, writer *tar.Writer, capabilties map[string][]string, labels map[string]string) error {
	for {
		entry, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := enrichEntry(entry, capabilties, labels); err != nil {
			return err
		}

		entry.Format = tar.FormatPAX
		if err := writer.WriteHeader(entry); err != nil {
			return err
		}
		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}
	}
	return nil
}

func enrichEntry(entry *tar.Header, capabilties map[string][]string, labels map[string]string) error {
	if entry.PAXRecords == nil {
		entry.PAXRecords = map[string]string{}
	}
	fileName := filepath.Clean(strings.TrimPrefix(entry.Name, "/"))

	if caps, exists := capabilties[fileName]; exists {
		if err := AddCapabilities(entry.PAXRecords, caps); err != nil {
			return err
		}
	}
	if l, exists := labels[fileName]; exists {
		if err := SetSELinuxLabel(entry.PAXRecords, l); err != nil {
			return err
		}
	}
	return nil
}
