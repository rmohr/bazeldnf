module github.com/rmohr/bazeldnf

go 1.20

require (
	github.com/bazelbuild/buildtools v0.0.0-20230127124510-cf446296fb76
	github.com/crillab/gophersat v1.3.1
	github.com/onsi/gomega v1.26.0
	github.com/sassoftware/go-rpmutils v0.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/klauspost/compress v1.11.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/sassoftware/go-rpmutils v0.2.0 => github.com/rmohr/go-rpmutils v0.1.2-0.20201215123907-5acf7436c00d
