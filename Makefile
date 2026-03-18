all: gazelle buildifier

deps-update:
	bazelisk run //:gazelle

gazelle:
	bazelisk run //:gazelle

test: gazelle buildifier e2e
	bazelisk build //... && bazelisk test //...

buildifier:
	bazelisk run //:buildifier.check

gofmt:
	gofmt -w pkg/.. cmd/..

e2e-workspace:
	bazel test e2e:workspace

e2e-bzlmod:
	bazel test e2e:bzlmod

e2e-bazel-bzlmod-lock-file:
	bazel test e2e:bzlmod-lock-file

e2e-bzlmod-build-toolchain:
	bazel test e2e:bzlmod-toolchain-from-source

e2e-bazel-bzlmod-lock-file-from-args:
	bazel test e2e:lock-file-from-args

e2e-bzlmod-toolchain-circular-dependencies:
	bazel test e2e:circular-deps

e2e:
	bazel test e2e

fmt: gofmt buildifier

.PHONY: gazelle test deps-update buildifier gofmt fmt e2e
