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
	@for version in 6.x 7.x; do \
		( \
			cd e2e/bazel-workspace && \
			echo "Testing $$version in workspace mode" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bzlmod:
	@for version in 7.x 8.x; do \
		( \
			cd e2e/bazel-bzlmod && \
			echo "Testing $$version with bzlmod" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bazel-bzlmod-lock-file:
	@for version in 7.x 8.x; do \
		( \
			cd e2e/bazel-bzlmod-lock-file && \
			echo "Testing $$version with bzlmod with lock file" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //...\
		) \
	done

e2e-bzlmod-build-toolchain:
	@for version in 7.x 8.x; do \
		( \
			cd e2e/bazel-bzlmod-toolchain-from-source && \
			echo "Testing $$version with bzlmod build toolchain" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //... --incompatible_enable_proto_toolchain_resolution \
		) \
	done

e2e-bazel-bzlmod-lock-file-from-args:
	@for version in 7.x 8.x; do \
		( \
			cd e2e/bazel-bzlmod-lock-file-from-args && \
			echo "Testing $$version bzlmod lock file from args" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //... --incompatible_enable_proto_toolchain_resolution \
		) \
	done

e2e-bzlmod-toolchain-circular-dependencies:
	@for version in 7.x 8.x; do \
		( \
			cd e2e/bzlmod-toolchain-circular-dependencies && \
			echo "Testing $$version bzlmod lock file from args" > /dev/stderr && \
			USE_BAZEL_VERSION=$$version bazelisk --batch build //... --incompatible_enable_proto_toolchain_resolution \
		) \
	done


e2e: e2e-workspace \
	e2e-bzlmod \
	e2e-bzlmod-build-toolchain \
	e2e-bazel-bzlmod-lock-file \
	e2e-bazel-bzlmod-lock-file-from-args \
	e2e-bzlmod-toolchain-circular-dependencies

fmt: gofmt buildifier

.PHONY: gazelle test deps-update buildifier gofmt fmt e2e
