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
	@for version in 5.x 6.x 7.x; do \
		( \
			cd e2e/bazel-workspace && \
			echo "Testing $$version" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bzlmod:
	@for version in 6.x 7.x; do \
		( \
			cd e2e/bazel-bzlmod && \
			echo "Testing $$version with bzlmod" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bzlmod-non-legacy-mode:
	@for version in 6.x 7.x; do \
		( \
			cd e2e/bazel-bzlmod-non-legacy-mode && \
			echo "Testing $$version with bzlmod with non-legacy mode" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bazel-bzlmod-lock-file:
	@for version in 6.x 7.x; do \
		( \
			cd e2e/bazel-bzlmod-lock-file && \
			echo "Testing $$version with bzlmod with lock file" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bzlmod-build-toolchain-6.x:
	( \
		cd e2e/bazel-bzlmod-toolchain-from-source && \
		USE_BAZEL_VERSION=6.x bazelisk --batch build //... \
	)

e2e-bzlmod-build-toolchain-7.x:
	( \
		cd e2e/bazel-bzlmod-toolchain-from-source && \
		USE_BAZEL_VERSION=7.x bazelisk --batch build //... --incompatible_enable_proto_toolchain_resolution \
	)

e2e-bzlmod-build-toolchain: e2e-bzlmod-build-toolchain-6.x e2e-bzlmod-build-toolchain-7.x

e2e: e2e-workspace e2e-bzlmod e2e-bzlmod-build-toolchain e2e-bzlmod-non-legacy-mode

fmt: gofmt buildifier

.PHONY: gazelle test deps-update buildifier gofmt fmt e2e
