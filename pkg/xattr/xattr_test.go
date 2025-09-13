package xattr

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

var g *GomegaWithT

func TestSettingSELinuxLabel(t *testing.T) {
	g = NewGomegaWithT(t)
	referenceEntry, err := getHeader("./selinux")
	g.Expect(err).ToNot(HaveOccurred())

	generatedEntry := &tar.Header{Name: "blub"}
	labels := map[string]string{"blub": "unconfined_u:object_r:user_home_t:s0", "somethingelse": "something"}

	g.Expect(enrichEntry(generatedEntry, nil, labels)).To(Succeed())

	g.Expect(generatedEntry.PAXRecords[selinux_header]).To(Equal(referenceEntry.PAXRecords[selinux_header]))
}

func getHeader(name string) (*tar.Header, error) {
	f, err := os.Open("testdata/xattr.tar")
	g.Expect(err).ToNot(HaveOccurred())
	defer f.Close()
	r := tar.NewReader(f)
	for {
		entry, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			g.Expect(err).ToNot(HaveOccurred())
		}
		if entry.Name == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("entry %v does not exist", name)
}

func Test_Capabilities(t *testing.T) {
	tests := []struct {
		name         string
		entry        string
		capabilities []string
	}{
		{
			name:         "should set cap_chown",
			entry:        "./cap_chown",
			capabilities: []string{"cap_chown"},
		},
		{
			name:         "should set cap_net_bind_service",
			entry:        "./cap_net_bind_service",
			capabilities: []string{"cap_net_bind_service"},
		},
		{
			name:         "should set cap_net_admin",
			entry:        "./cap_net_admin", 
			capabilities: []string{"cap_net_admin"},
		},
		{
			name:         "should set cap_net_raw",
			entry:        "./cap_net_raw",
			capabilities: []string{"cap_net_raw"},
		},
		{
			name:         "should set cap_sys_ptrace",
			entry:        "./cap_sys_ptrace",
			capabilities: []string{"cap_sys_ptrace"},
		},
		{
			name:         "should set all implemented capabilities",
			entry:        "./cap_all",
			capabilities: []string{"cap_net_bind_service", "cap_net_admin", "cap_net_raw", "cap_chown", "cap_sys_ptrace"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g = NewGomegaWithT(t)
			referenceEntry, err := getHeader(tt.entry)
			g.Expect(err).ToNot(HaveOccurred())

			generatedEntry := &tar.Header{Name: "blub"}

			g.Expect(enrichEntry(generatedEntry, map[string][]string{"blub": tt.capabilities}, nil)).To(Succeed())

			g.Expect(generatedEntry.PAXRecords[capabilities_header]).To(Equal(referenceEntry.PAXRecords[capabilities_header]))
		})
	}
}
