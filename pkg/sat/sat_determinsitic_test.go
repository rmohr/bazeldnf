package sat

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
	"github.com/rmohr/bazeldnf/pkg/api/bazeldnf"
)

func TestDeterministicOutput(t *testing.T) {
	tt := struct {
		requires []string
		installs []string
		repofile string
	}{
		requires: []string{
			"shadow-utils",
		},
		installs: []string{
			"glibc-all-langpacks-0:2.28-161.el8.x86_64",
			"libsepol-0:2.9-2.el8.x86_64",
			"ncurses-base-0:6.1-9.20180224.el8.noarch",
			"libacl-0:2.2.53-1.el8.x86_64",
			"libattr-0:2.4.48-3.el8.x86_64",
			"basesystem-0:11-5.el8.noarch",
			"bash-0:4.4.20-1.el8_4.x86_64",
			"filesystem-0:3.8-6.el8.x86_64",
			"centos-gpg-keys-1:8-2.el8.noarch",
			"libsemanage-0:2.9-6.el8.x86_64",
			"libxcrypt-0:4.1.1-6.el8.x86_64",
			"libcap-0:2.26-4.el8.x86_64",
			"libselinux-0:2.9-5.el8.x86_64",
			"ncurses-libs-0:6.1-9.20180224.el8.x86_64",
			"bzip2-libs-0:1.0.6-26.el8.x86_64",
			"coreutils-single-0:8.30-10.el8.x86_64",
			"tzdata-0:2021a-1.el8.noarch",
			"shadow-utils-2:4.6-13.el8.x86_64",
			"glibc-0:2.28-161.el8.x86_64",
			"glibc-common-0:2.28-161.el8.x86_64",
			"libcap-ng-0:0.7.11-1.el8.x86_64",
			"pcre2-0:10.32-2.el8.x86_64",
			"audit-libs-0:3.0-0.17.20191104git1c2f876.el8.x86_64",
			"setup-0:2.12.2-6.el8.noarch",
			"centos-stream-release-0:8.5-3.el8.noarch",
			"centos-stream-repos-0:8-2.el8.noarch",
		},
		repofile: "testdata/libguestfs-el8.xml",
	}

	for i := range 10 {
		t.Run(fmt.Sprintf("iteration: %d", i), func(t *testing.T) {
			g := NewGomegaWithT(t)
			f, err := os.Open(tt.repofile)
			g.Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			repo := &api.Repository{}
			err = xml.NewDecoder(f).Decode(repo)
			g.Expect(err).ToNot(HaveOccurred())

			packages := []*api.Package{}
			for i, _ := range repo.Packages {
				pkg := &repo.Packages[i]
				pkg.Repository = &bazeldnf.Repository{}
				packages = append(packages, pkg)
			}

			loader := NewLoader()
			model, err := loader.Load(packages, tt.requires, nil, nil, false, []string{"x86_64", "noarch"})
			g.Expect(err).ToNot(HaveOccurred())

			install, _, _, err := Resolve(model)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(installToString(install)).To(ConsistOf(tt.installs))
		})
	}

}

func installToString(given []*api.Package) (resolved []string) {
	for _, p := range given {
		resolved = append(resolved, p.String())
	}
	return
}
