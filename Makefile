all: gazelle buildifier

deps-update:
	bazelisk run //:gazelle

gazelle:
	bazelisk run //:gazelle

test: gazelle e2e
	bazelisk build //... && bazelisk test //...

buildifier:
	bazelisk run //:buildifier.check

gofmt:
	gofmt -w pkg/.. cmd/..

e2e:
	@for version in 5.x 6.x 7.x; do \
		( \
			cd e2e/bazel-workspace && \
			echo "Testing $$version" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

fmt: gofmt buildifier

.PHONY: gazelle test deps-update buildifier gofmt fmt e2e
