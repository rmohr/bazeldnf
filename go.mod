module github.com/rmohr/bazeldnf

go 1.14

require (
	github.com/bazelbuild/buildtools v0.0.0-20221004120235-7186f635531b
	github.com/crillab/gophersat v1.3.1
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/onsi/gomega v1.10.3
	github.com/sassoftware/go-rpmutils v0.1.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.0
	github.com/spf13/viper v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
	golang.org/x/sys v0.1.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace github.com/sassoftware/go-rpmutils v0.1.1 => github.com/rmohr/go-rpmutils v0.1.2-0.20201215123907-5acf7436c00d
