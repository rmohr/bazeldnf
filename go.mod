module github.com/rmohr/bazeldnf

go 1.14

require (
	github.com/bazelbuild/buildtools v0.0.0-20201023142455-8a8e1e724705
	github.com/crillab/gophersat v1.3.1
	github.com/onsi/gomega v1.10.3
	github.com/sassoftware/go-rpmutils v0.1.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/sassoftware/go-rpmutils v0.1.1 => github.com/rmohr/go-rpmutils v0.1.2-0.20201215123907-5acf7436c00d
