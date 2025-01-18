package sat

import (
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rmohr/bazeldnf/pkg/api"
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
			"glibc-all-langpacks-0:2.28-161.el8",
			"libsepol-0:2.9-2.el8",
			"ncurses-base-0:6.1-9.20180224.el8",
			"libacl-0:2.2.53-1.el8",
			"libattr-0:2.4.48-3.el8",
			"basesystem-0:11-5.el8",
			"bash-0:4.4.20-1.el8_4",
			"filesystem-0:3.8-6.el8",
			"centos-gpg-keys-1:8-2.el8",
			"libsemanage-0:2.9-6.el8",
			"libxcrypt-0:4.1.1-6.el8",
			"libcap-0:2.26-4.el8",
			"libselinux-0:2.9-5.el8",
			"ncurses-libs-0:6.1-9.20180224.el8",
			"bzip2-libs-0:1.0.6-26.el8",
			"coreutils-single-0:8.30-10.el8",
			"tzdata-0:2021a-1.el8",
			"shadow-utils-2:4.6-13.el8",
			"glibc-0:2.28-161.el8",
			"glibc-common-0:2.28-161.el8",
			"libcap-ng-0:0.7.11-1.el8",
			"pcre2-0:10.32-2.el8",
			"audit-libs-0:3.0-0.17.20191104git1c2f876.el8",
			"setup-0:2.12.2-6.el8",
			"centos-stream-release-0:8.5-3.el8",
			"centos-stream-repos-0:8-2.el8",
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

			resolver := NewResolver(false)
			packages := []*api.Package{}
			for i, _ := range repo.Packages {
				packages = append(packages, &repo.Packages[i])
			}
			err = resolver.LoadInvolvedPackages(packages, nil, nil)
			g.Expect(err).ToNot(HaveOccurred())
			err = resolver.ConstructRequirements(tt.requires)
			g.Expect(err).ToNot(HaveOccurred())
			install, _, _, err := resolver.Resolve()
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
