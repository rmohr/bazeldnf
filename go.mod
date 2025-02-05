module github.com/rmohr/bazeldnf

go 1.22.0

toolchain go1.22.2

require (
	github.com/bazelbuild/buildtools v0.0.0-20250110114635-13fa61383b99
	github.com/crillab/gophersat v1.4.0
	github.com/onsi/gomega v1.36.2
	github.com/sassoftware/go-rpmutils v0.2.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	golang.org/x/crypto v0.32.0
	sigs.k8s.io/yaml v1.4.0
)

require github.com/bazelbuild/rules_go v0.52.0

require (
	github.com/adrg/xdg v0.5.3
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.11.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/sassoftware/go-rpmutils v0.2.0 => github.com/rmohr/go-rpmutils v0.1.2-0.20201215123907-5acf7436c00d
