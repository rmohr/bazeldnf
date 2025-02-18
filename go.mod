module github.com/rmohr/bazeldnf

go 1.22.2

toolchain go1.22.7

require (
	github.com/bazelbuild/buildtools v0.0.0-20250110114635-13fa61383b99
	github.com/crillab/gophersat v1.4.0
	github.com/onsi/gomega v1.36.2
	github.com/sassoftware/go-rpmutils v0.2.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	golang.org/x/crypto v0.33.0
	sigs.k8s.io/yaml v1.4.0
)

require github.com/bazelbuild/rules_go v0.52.0

require (
	github.com/STARRY-S/zip v0.2.1 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.0 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/nwaples/rardecode/v2 v2.0.0-beta.4.0.20241112120701-034e449c6e78 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/sorairolake/lzip-go v0.3.5 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
)

require (
	github.com/adrg/xdg v0.5.3
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mholt/archives v0.1.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	golang.org/x/net v0.35.0
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/sassoftware/go-rpmutils v0.2.0 => github.com/rmohr/go-rpmutils v0.1.2-0.20201215123907-5acf7436c00d
