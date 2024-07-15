all: gazelle buildifier

deps-update:
	bazelisk run //:gazelle

gazelle:
	bazelisk run //:gazelle

test: gazelle
	bazelisk build //... && bazelisk test //...

buildifier:
	bazelisk run //:buildifier

gofmt:
	gofmt -w pkg/.. cmd/..

e2e:
	(cd e2e/bazel-5 && bazelisk build //...)
	(cd e2e/bazel-6 && bazelisk build //...)
	(cd e2e/bazel-6-bzlmod && bazelisk build //...)
	(cd e2e/bazel-7 && bazelisk build //...)
	(cd e2e/bazel-7-bzlmod && bazelisk build //...)

fmt: gofmt buildifier

.PHONY: gazelle test deps-update buildifier gofmt fmt e2e
