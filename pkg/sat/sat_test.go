package sat

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazel-dnf/pkg/api"
)

func Test(t *testing.T) {
	tests := []struct {
		name     string
		requires []string
		installs []string
		repofile string
	}{
		{name: "should resolve bash",
			requires: []string{
				"bash",
				"fedora-release-server",
				"glibc-langpack-en",
			},
			installs: []string{
				"libgcc-0:10.2.1-1.fc32",
				"fedora-gpg-keys-0:32-6",
				"glibc-0:2.31-4.fc32",
				"glibc-langpack-en-0:2.31-4.fc32",
				"fedora-release-common-0:32-3",
				"glibc-common-0:2.31-4.fc32",
				"ncurses-base-0:6.1-15.20191109.fc32",
				"ncurses-libs-0:6.1-15.20191109.fc32",
				"fedora-release-server-0:32-3",
				"tzdata-0:2020a-1.fc32",
				"setup-0:2.13.6-2.fc32",
				"basesystem-0:11-9.fc32",
				"bash-0:5.0.17-1.fc32",
				"filesystem-0:3.14-2.fc32",
				"fedora-repos-0:32-6",
			},
			repofile: "../../testdata/bash-fc31.xml",
		},
		{name: "should resolve libvirt-daemon",
			requires: []string{
				"libvirt-daemon",
				"fedora-release-server",
				"glibc-langpack-en",
				"coreutils-single",
				"libcurl-minimal",
			},
			installs: []string{
				"keyutils-libs-0:1.6-4.fc32",
				"filesystem-0:3.14-2.fc32",
				"iproute-tc-0:5.7.0-1.fc32",
				"zlib-0:1.2.11-21.fc32",
				"bash-0:5.0.17-1.fc32",
				"libuuid-0:2.35.2-1.fc32",
				"qrencode-libs-0:4.0.2-5.fc32",
				"libcom_err-0:1.45.5-3.fc32",
				"lz4-libs-0:1.9.1-2.fc32",
				"alternatives-0:1.11-6.fc32",
				"openssl-libs-1:1.1.1g-1.fc32",
				"expat-0:2.2.8-2.fc32",
				"libssh-config-0:0.9.5-1.fc32",
				"libgpg-error-0:1.36-3.fc32",
				"libattr-0:2.4.48-8.fc32",
				"ncurses-base-0:6.1-15.20191109.fc32",
				"libselinux-0:3.0-5.fc32",
				"libunistring-0:0.9.10-7.fc32",
				"libnl3-0:3.5.0-2.fc32",
				"libpcap-14:1.9.1-3.fc32",
				"gnutls-0:3.6.15-1.fc32",
				"libxml2-0:2.9.10-7.fc32",
				"setup-0:2.13.6-2.fc32",
				"libffi-0:3.1-24.fc32",
				"libmount-0:2.35.2-1.fc32",
				"libvirt-libs-0:6.1.0-4.fc32",
				"dbus-libs-1:1.12.20-1.fc32",
				"ca-certificates-0:2020.2.41-1.1.fc32",
				"psmisc-0:23.3-3.fc32",
				"xz-libs-0:5.2.5-1.fc32",
				"iproute-0:5.7.0-1.fc32",
				"openldap-0:2.4.47-5.fc32",
				"libidn2-0:2.3.0-2.fc32",
				"acl-0:2.2.53-5.fc32",
				"libacl-0:2.2.53-5.fc32",
				"libgcc-0:10.2.1-1.fc32",
				"libmnl-0:1.0.4-11.fc32",
				"libtirpc-0:1.2.6-1.rc4.fc32",
				"libargon2-0:20171227-4.fc32",
				"libcap-ng-0:0.7.11-1.fc32",
				"device-mapper-0:1.02.171-1.fc32",
				"mpfr-0:4.0.2-5.fc32",
				"libcurl-minimal-0:7.69.1-6.fc32",
				"linux-atm-libs-0:2.5.1-26.fc32",
				"systemd-pam-0:245.8-2.fc32",
				"sed-0:4.5-5.fc32",
				"systemd-libs-0:245.8-2.fc32",
				"gzip-0:1.10-2.fc32",
				"krb5-libs-0:1.18.2-22.fc32",
				"crypto-policies-0:20200619-1.git781bbd4.fc32",
				"dbus-common-1:1.12.20-1.fc32",
				"nettle-0:3.5.1-5.fc32",
				"numactl-libs-0:2.0.12-4.fc32",
				"libtasn1-0:4.16.0-1.fc32",
				"libsigsegv-0:2.11-10.fc32",
				"libsepol-0:3.0-4.fc32",
				"polkit-0:0.116-7.fc32",
				"json-c-0:0.13.1-13.fc32",
				"glibc-common-0:2.31-4.fc32",
				"dbus-broker-0:24-1.fc32",
				"audit-libs-0:3.0-0.19.20191104git1c2f876.fc32",
				"nmap-ncat-2:7.80-4.fc32",
				"tzdata-0:2020a-1.fc32",
				"device-mapper-libs-0:1.02.171-1.fc32",
				"shadow-utils-2:4.8.1-2.fc32",
				"fedora-repos-0:32-6",
				"libfdisk-0:2.35.2-1.fc32",
				"pcre2-0:10.35-6.fc32",
				"fedora-release-common-0:32-3",
				"libsmartcols-0:2.35.2-1.fc32",
				"dmidecode-1:3.2-5.fc32",
				"pcre2-syntax-0:10.35-6.fc32",
				"bzip2-libs-0:1.0.8-2.fc32",
				"libstdc++-0:10.2.1-1.fc32",
				"libnghttp2-0:1.41.0-1.fc32",
				"libxcrypt-0:4.4.17-1.fc32",
				"libdb-0:5.3.28-40.fc32",
				"glib2-0:2.64.5-1.fc32",
				"elfutils-default-yama-scope-0:0.181-1.fc32",
				"systemd-rpm-macros-0:245.8-2.fc32",
				"libnsl2-0:1.2.0-6.20180605git4a062cf.fc32",
				"glibc-0:2.31-4.fc32",
				"libsemanage-0:3.0-3.fc32",
				"libgcrypt-0:1.8.5-3.fc32",
				"systemd-0:245.8-2.fc32",
				"libvirt-daemon-0:6.1.0-4.fc32",
				"polkit-libs-0:0.116-7.fc32",
				"p11-kit-trust-0:0.23.21-2.fc32",
				"fedora-gpg-keys-0:32-6",
				"iptables-libs-0:1.8.4-9.fc32",
				"kmod-libs-0:27-1.fc32",
				"libutempter-0:1.1.6-18.fc32",
				"libcap-0:2.26-7.fc32",
				"ncurses-libs-0:6.1-15.20191109.fc32",
				"gmp-1:6.1.2-13.fc32",
				"cyrus-sasl-0:2.1.27-4.fc32",
				"libverto-0:0.3.0-9.fc32",
				"readline-0:8.0-4.fc32",
				"polkit-pkla-compat-0:0.1-16.fc32",
				"libnetfilter_conntrack-0:1.0.7-4.fc32",
				"coreutils-single-0:8.32-4.fc32.1",
				"libssh-0:0.9.5-1.fc32",
				"kmod-0:27-1.fc32",
				"util-linux-0:2.35.2-1.fc32",
				"libseccomp-0:2.5.0-3.fc32",
				"grep-0:3.3-4.fc32",
				"glibc-langpack-en-0:2.31-4.fc32",
				"p11-kit-0:0.23.21-2.fc32",
				"libblkid-0:2.35.2-1.fc32",
				"libwsman1-0:2.6.8-12.fc32",
				"cryptsetup-libs-0:2.3.4-1.fc32",
				"mozjs60-0:60.9.0-5.fc32",
				"elfutils-libelf-0:0.181-1.fc32",
				"libpwquality-0:1.4.2-2.fc32",
				"fedora-release-server-0:32-3",
				"cyrus-sasl-gssapi-0:2.1.27-4.fc32",
				"gawk-0:5.0.1-7.fc32",
				"basesystem-0:11-9.fc32",
				"numad-0:0.5-31.20150602git.fc32",
				"libssh2-0:1.9.0-5.fc32",
				"elfutils-libs-0:0.181-1.fc32",
				"dbus-1:1.12.20-1.fc32",
				"yajl-0:2.1.0-14.fc32",
				"pam-0:1.3.1-26.fc32",
				"cyrus-sasl-lib-0:2.1.27-4.fc32",
				"libnfnetlink-0:1.0.1-17.fc32",
				"pcre-0:8.44-1.fc32",
				"cracklib-0:2.9.6-22.fc32",
			},
			repofile: "../../testdata/libvirt-daemon-fc31.xml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			f, err := os.Open(tt.repofile)
			g.Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			repo := &api.Repository{}
			err = xml.NewDecoder(f).Decode(repo)
			g.Expect(err).ToNot(HaveOccurred())

			resolver := NewResolver(false)
			packages := []*api.Package{}
			for i, _ := range repo.Packages {
				packages = append(packages, &repo.Packages[i])
			}
			err = resolver.LoadInvolvedPackages(packages)
			g.Expect(err).ToNot(HaveOccurred())
			err = resolver.ConstructRequirements(tt.requires)
			g.Expect(err).ToNot(HaveOccurred())
			install, _, err := resolver.Resolve()
			g.Expect(pkgToString(install)).To(ConsistOf(tt.installs))
			g.Expect(err).ToNot(HaveOccurred())
		})
	}
}

func TestNewResolver(t *testing.T) {
	tests := []struct {
		name     string
		packages []*api.Package
		requires []string
		install  []string
		exclude  []string
		solvable bool
	}{
		{name: "with indirect dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d", "g"}),
			newPkg("testb", "1", []string{"testb", "c"}, []string{}),
			newPkg("testc", "1", []string{"testc", "d"}, []string{}),
			newPkg("testd", "1", []string{"testd", "e", "f", "g"}, []string{"h"}),
			newPkg("teste", "1", []string{"teste", "h"}, []string{}),
		}, requires: []string{
			"testa",
		},
			install:  []string{"testa-0:1", "testc-0:1", "testd-0:1", "teste-0:1"},
			exclude:  []string{"testb-0:1"},
			solvable: true,
		},
		{name: "with circular dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d", "g"}),
			newPkg("testb", "1", []string{"testb", "c"}, []string{}),
			newPkg("testc", "1", []string{"testc", "d"}, []string{}),
			newPkg("testd", "1", []string{"testd", "e", "f", "g"}, []string{"h"}),
			newPkg("teste", "1", []string{"teste", "h"}, []string{"a"}),
		}, requires: []string{
			"testa",
		},
			install:  []string{"testa-0:1", "testc-0:1", "testd-0:1", "teste-0:1"},
			exclude:  []string{"testb-0:1"},
			solvable: true,
		},
		{name: "with an unresolvable dependency", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d"}),
		}, requires: []string{
			"testa",
		},
			solvable: false,
		},
		{name: "with two sources to choose from, should use the newer one", packages: []*api.Package{
			newPkg("testa", "1", []string{"testa", "a", "b"}, []string{"d"}),
			newPkg("testb", "1", []string{"testb", "d"}, []string{}),
			newPkg("testb", "2", []string{"testb", "d"}, []string{}),
		}, requires: []string{
			"testa",
		},
			install:  []string{"testa-0:1", "testb-0:2"},
			exclude:  []string{},
			solvable: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(false)
			err := resolver.LoadInvolvedPackages(tt.packages)
			if err != nil {
				t.Fail()
			}
			err = resolver.ConstructRequirements(tt.requires)
			if err != nil {
				fmt.Println(err)
				t.Fail()
			}
			install, exclude, err := resolver.Resolve()
			g := NewGomegaWithT(t)
			if tt.solvable {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(pkgToString(install)).To(ConsistOf(tt.install))
			g.Expect(pkgToString(exclude)).To(ConsistOf(tt.exclude))
		})
	}
}

func newPkg(name string, version string, provides []string, requires []string) *api.Package {
	pkg := &api.Package{}
	pkg.Name = name
	pkg.Version = api.Version{Ver: version}
	for _, req := range requires {
		pkg.Format.Requires.Entries = append(pkg.Format.Requires.Entries, api.Entry{Name: req})
	}
	pkg.Format.Provides.Entries = append(pkg.Format.Provides.Entries, api.Entry{
		Name:  name,
		Flags: "EQ",
		Ver:   version,
	})
	for _, req := range provides {
		pkg.Format.Provides.Entries = append(pkg.Format.Provides.Entries, api.Entry{Name: req})
	}

	return pkg
}

func strToPkg(wanted []string, given []*api.Package) (resolved []*api.Package) {
	m := map[string]*api.Package{}
	for _, p := range given {
		m[p.String()] = p
	}
	for _, w := range wanted {
		resolved = append(resolved, m[w])
	}
	return resolved
}

func pkgToString(given []*api.Package) (resolved []string) {
	for _, p := range given {
		resolved = append(resolved, p.String())
	}
	return
}
