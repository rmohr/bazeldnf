module github.com/rmohr/bazeldnf

go 1.14

require (
	github.com/crillab/gophersat v1.3.1
	github.com/onsi/gomega v1.10.3
	github.com/sassoftware/go-rpmutils v0.1.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
)

replace github.com/sassoftware/go-rpmutils v0.1.1 => github.com/rmohr/go-rpmutils v0.1.2-0.20201019160740-3eae840fd906
