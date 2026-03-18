# Contributing to bazeldnf

Thank you for your interest in contributing to bazeldnf! This guide covers the
development workflow for making changes to Go code, Starlark rules, dependency
management, and end-to-end tests.

## Prerequisites

- [Bazelisk](https://github.com/bazelbuild/bazelisk) (symlinked as `bazel`)
- Git

## Project Layout

```
bazeldnf/
├── cmd/                    # CLI binary (cobra commands)
├── pkg/                    # Go library packages
│   ├── api/                # Core API types
│   ├── bazel/              # BUILD file manipulation
│   ├── ldd/                # Shared library dependency analysis
│   ├── order/              # Package ordering
│   ├── reducer/            # SAT solver wrapper
│   ├── repo/               # DNF/RPM repo fetching and parsing
│   ├── rpm/                # RPM file handling
│   ├── sat/                # SAT solver integration
│   └── xattr/              # Extended attributes
├── bazeldnf/               # Public Starlark API (.bzl files)
│   └── private/            # Private Starlark helpers
├── internal/               # Starlark rule implementations
├── tools/                  # Release tooling, version info, integrity checksums
├── e2e/                    # End-to-end integration test workspaces
├── MODULE.bazel            # Bazel module definition (bzlmod)
├── go.mod / go.sum         # Go module dependencies
└── BUILD.bazel             # Root build targets (gazelle, buildifier, etc.)
```

## Building and Testing

```sh
# Build everything
bazel build //...

# Run all unit tests
bazel test //...

# Run all end-to-end tests that should pass
bazel test e2e

# Run end-to-end tests that are allowed to fail
bazel test //e2e/... --test_tag_filters=allowed-to-fail
```

## Making Changes to Go Code

The Go source lives in `cmd/` (CLI binary) and `pkg/` (library packages).

### Workflow

1. **Make your code changes** in the relevant `cmd/` or `pkg/` package.

2. **Format your code**:
   ```sh
   bazel run @rules_go//go -- fmt ./...
   ```

3. **Regenerate BUILD files** with Gazelle. Bazel BUILD files for Go packages
   are managed by Gazelle, so any new files, packages, or import changes
   require updating them:
   ```sh
   bazel run //:gazelle
   ```

4. **Run unit tests** to verify your changes:
   ```sh
   bazel test //...
   ```

### Adding a New Go Package

If you add a new package under `pkg/` or `cmd/`:

1. Create the directory and add your `.go` files.
2. Run Gazelle to auto-generate the `BUILD.bazel` file:
   ```sh
   bazel run //:gazelle
   ```
   Gazelle will create the `go_library`, `go_test`, and/or `go_binary` targets
   automatically based on your source files.

### Unit Tests

Go tests use [gomega](https://onsi.github.io/gomega/) for assertions. Place
test files (`*_test.go`) alongside the code they test. Gazelle will
automatically include them in the generated `go_test` targets.

## Adding New Go Dependencies

Go dependencies are managed with standard Go modules and synced to Bazel via
Gazelle and bzlmod.

1. **Add the dependency**:
   ```sh
   bazel run @rules_go//go -- get github.com/example/newdep@latest
   ```
   This updates `go.mod` and `go.sum`.

2. **Sync bzlmod `use_repo` declarations**:
   ```sh
   bazel mod tidy
   ```
   This automatically updates the `use_repo` calls in `MODULE.bazel`.

3. **Regenerate BUILD files**:
   ```sh
   bazel run //:gazelle
   ```

4. **Verify the build**:
   ```sh
   bazel build //...
   bazel test //...
   ```

## Adding New Bazel Dependencies

Bazel module dependencies are managed in `MODULE.bazel` using bzlmod.

1. **Add the `bazel_dep`** to `MODULE.bazel`:
   ```python
   # Use dev_dependency = True if only needed for development/testing
   bazel_dep(name = "rules_foo", version = "1.2.3", dev_dependency = True)
   ```

2. **Sync `use_repo` declarations**:
   ```sh
   bazel mod tidy
   ```

3. **Run Gazelle** if the new dependency provides Starlark libraries that
   Gazelle should be aware of:
   ```sh
   bazel run //:gazelle
   ```

4. **Verify everything builds**:
   ```sh
   bazel build //...
   ```

### Dependency Categories

In `MODULE.bazel`, dependencies are organized as:

- **Runtime dependencies** (not `dev_dependency`): required by users of
  bazeldnf (e.g., `bazel_skylib`, `platforms`, `bazel_features`, `bazel_lib`).
  Be conservative adding non-dev dependencies as they affect all consumers.
- **Toolchain build dependencies**: `gazelle` and `rules_go`, needed when
  building the toolchain from source.
- **Dev dependencies** (`dev_dependency = True`): only used within this repo
  for building, testing, and linting (e.g., `buildifier_prebuilt`,
  `rules_bazel_integration_test`).

## Making Starlark Changes

Starlark (`.bzl`) files define the Bazel rules, macros, and extensions that
users interact with.

### File Organization

| Directory | Purpose |
|---|---|
| `bazeldnf/defs.bzl` | Public API facade (re-exports all user-facing rules) |
| `bazeldnf/extensions.bzl` | Bzlmod module extensions (`bazeldnf_toolchain`, `bazeldnf`) |
| `bazeldnf/toolchain.bzl` | Toolchain rule definition |
| `bazeldnf/repositories.bzl` | WORKSPACE-mode setup functions |
| `bazeldnf/private/` | Private implementation details (lock file helpers, repo rules) |
| `internal/` | Rule implementations (`rpm.bzl`, `rpmtree.bzl`, `bazeldnf.bzl`, `xattrs.bzl`) |

### Workflow

1. **Make your `.bzl` changes**.

2. **Format with buildifier**:
   ```sh
   bazel run //:buildifier
   ```

3. **Run Gazelle** to update any `bzl_library` targets if you added or renamed
   `.bzl` files:
   ```sh
   bazel run //:gazelle
   ```

4. **Run unit tests and e2e tests** to verify nothing is broken:
   ```sh
   bazel test //...
   bazel test e2e
   ```

### Tips

- When modifying user-facing rules in `internal/`, update the re-exports in
  `bazeldnf/defs.bzl` if adding new symbols.
- If you change the module extension in `bazeldnf/extensions.bzl`, test with
  both the lock-file and non-lock-file e2e workspaces.
- The `internal/runner.bash.template` is used to generate runner scripts for
  the `bazeldnf` runnable rule.

## End-to-End Tests

E2E tests use
[rules_bazel_integration_test](https://github.com/bazel-contrib/rules_bazel_integration_test)
to spin up isolated Bazel workspaces and verify that bazeldnf works correctly
across different Bazel versions and configurations.

### Existing E2E Test Suites

| Suite | Directory | Description |
|---|---|---|
| `e2e:bzlmod` | `e2e/bazel-bzlmod/` | Basic bzlmod with prebuilt toolchain |
| `e2e:bzlmod-lock-file` | `e2e/bazel-bzlmod-lock-file/` | Bzlmod with lock file workflow |
| `e2e:bzlmod-toolchain-from-source` | `e2e/bazel-bzlmod-toolchain-from-source/` | Toolchain built from Go source |
| `e2e:lock-file-from-args` | `e2e/bazel-bzlmod-lock-file-from-args/` | Lock file generated from CLI args |
| `e2e:circular-deps` | `e2e/bzlmod-toolchain-circular-dependencies/` | Circular RPM dependency handling |
| `e2e:repo-yaml` | `e2e/repo-yaml/` | CLI `init`/`fetch`/`resolve` workflow |

### Running E2E Tests

```sh
# Run all e2e tests
bazel test e2e

# Run a specific suite
bazel test e2e:bzlmod
bazel test e2e:bzlmod-lock-file
```

### Adding a New E2E Test

#### 1. Create the Workspace Directory

Create a new directory under `e2e/` with the files for your test workspace:

```
e2e/my-new-test/
├── MODULE.bazel    # Module definition, references parent
├── BUILD.bazel     # Build targets that exercise your scenario
└── .bazelrc        # Required, see below
```

#### 2. Update `.bazelignore` and Deleted Packages

Since e2e workspaces are nested Bazel projects, you must prevent the parent
workspace from traversing into their Bazel output directories, and update the
`--deleted_packages` flags so the parent doesn't treat child `BUILD` files as
its own packages.

1. **Add `bazel-*` output symlinks to `.bazelignore`**. Append entries for your
   new workspace (follow the pattern of existing entries):
   ```
   e2e/my-new-test/bazel-bin
   e2e/my-new-test/bazel-out
   e2e/my-new-test/bazel-my-new-test
   e2e/my-new-test/bazel-testlogs
   ```

2. **Regenerate `--deleted_packages` flags**:
   ```sh
   bazel run @rules_bazel_integration_test//tools:update_deleted_packages
   ```
   This updates the `.bazelrc` files with the correct `--deleted_packages`
   entries so the parent workspace ignores `BUILD` files inside child
   workspaces.

#### 3. Create `.bazelrc`

Every e2e workspace needs a `.bazelrc` that imports shared configs from the
parent repository:

```
# Import Aspect bazelrc presets
import %workspace%/../../.aspect/bazelrc/bazel7.bazelrc

# Include our e2e shared config
import %workspace%/../.bazelrc

# Specific project flags go here if needed

# Load any settings & overrides specific to the current user from `.bazelrc.user`.
# This file should appear in `.gitignore` so that settings are not shared with team members. This
# should be last statement in this config so the user configuration is able to overwrite flags from
# this file. See https://bazel.build/configure/best-practices#bazelrc-file.
try-import %workspace%/../../.bazelrc.user
```

#### 4. Set Up MODULE.bazel

Your test workspace must reference the parent repository using
`local_path_override` so that it uses the local (in-development) version of
bazeldnf:

```python
module(name = "my-new-test")

bazel_dep(name = "bazeldnf")
local_path_override(
    module_name = "bazeldnf",
    path = "../..",
)

# Use bazeldnf extensions as needed
bazeldnf = use_extension("@bazeldnf//bazeldnf:extensions.bzl", "bazeldnf")
# ... configure RPMs, lock files, etc.
```

#### 5. Create BUILD Targets

In your test workspace's `BUILD.bazel`, use the bazeldnf rules to set up the
scenario you want to test:

```python
load("@bazeldnf//bazeldnf:defs.bzl", "bazeldnf", "rpmtree", "tar2files")

bazeldnf(name = "bazeldnf")

rpmtree(
    name = "my_rpms",
    rpms = ["@some-rpm//rpm"],
)
```

#### 6. Register the Test in `e2e/BUILD.bazel`

Add the integration test definition to `e2e/BUILD.bazel`. Use the existing
`:test-runner` which runs `bazel --nohome_rc build //...` inside the workspace:

```python
bazel_integration_tests(
    name = "e2e_my_new_test",
    timeout = DEFAULT_TIMEOUT,
    bazel_binaries = bazel_binaries,
    bazel_versions = BZLMOD_BAZEL_VERSIONS,
    tags = DEFAULT_TEST_TAGS,
    test_runner = ":test-runner",
    workspace_files = glob_workspace_files("my-new-test") + [
        ".bazelrc",
        "//:local_repository_files",
    ],
    workspace_path = "my-new-test",
)
```

If you're developing a new feature which isn't fully ready yet, but it doesn't
add regressions, then you can add `['allowed-to-fail']` to your tags, this will
allow your new feature to be merged incrementally and making the reviewing
process easier.

If your test needs to run commands beyond `build //...` (e.g., running
`fetch` or `resolve` before building), define a custom test runner:

```python
default_test_runner(
    name = "my-test-runner",
    bazel_cmds = [
        "run :bazeldnf -- fetch --cache-dir $(pwd)/.bazeldnf",
        "build //...",
    ],
)
```

#### 7. Add a Test Suite

Add a test suite so the test can be run by name:

```python
test_suite(
    name = "my-new-test",
    tags = DEFAULT_TEST_SUITE_TAGS,
    tests = integration_test_utils.bazel_integration_test_names(
        "e2e_my_new_test",
        BZLMOD_BAZEL_VERSIONS,
    ),
)
```

In the case you needed to add `allowed-to-fail` to the test tags also add it here.

And include it in the top-level `e2e` test suite:

```python
test_suite(
    name = "e2e",
    tags = DEFAULT_TEST_SUITE_TAGS,
    tests = [
        # ... existing suites ...
        ":my-new-test",
    ],
)
```

#### 8. Expose Parent Files

If your test workspace depends on parent files not already included in
`//:local_repository_files` (defined in the root `BUILD.bazel`), add them to
that filegroup. Each package that needs to be visible must have:

```python
filegroup(
    name = "all_files",
    srcs = glob(["**"]),
    visibility = ["//:__pkg__"],
)
```

#### 9. Verify Everything

Run your new test to make sure it passes:

```sh
bazel test e2e:my-new-test
```

In the case your tests are allowed to fail then you do

```sh
bazel test e2e:my-new-test --test_tag_filters=allowed-to-fail
```

### Key Concepts for E2E Tests

- **`local_path_override`**: each child workspace references the parent repo at
  `../..`, so it always tests the local development version of bazeldnf.
- **`//:local_repository_files`**: a filegroup in the root `BUILD.bazel` that
  aggregates all parent workspace files needed by e2e workspaces. The
  integration test framework copies these into the test sandbox.
- **`glob_workspace_files`**: a helper in `e2e/helpers.bzl` that globs all
  files in a workspace directory, excluding Bazel output directories
  (`bazel-*`) and `MODULE.bazel.lock` files.
- **`--nohome_rc`**: all test runners automatically prepend this flag to
  prevent the host's `~/.bazelrc` from interfering with tests.
- **Bazel versions**: bzlmod tests run against `BZLMOD_BAZEL_VERSIONS` (7.6.0,
  8.1.0).

## Formatting and Linting

All formatting checks run in CI. Make sure they pass before submitting:

```sh
# Check Starlark/BUILD formatting (CI runs this)
bazel run //:buildifier.check

# Check BUILD files are up-to-date (CI runs this)
bazel run //:gazelle.check

# Auto-fix Starlark formatting
bazel run //:buildifier

# Regenerate BUILD files
bazel run //:gazelle

# Format Go code
bazel run @rules_go//go -- fmt ./...
```

## CI

GitHub Actions runs on every push and PR to `main`:

1. **Linter** (`linter.yaml`): checks buildifier and gazelle formatting.
2. **Main CI** (`action.yml`): builds everything, runs unit tests, then runs
   all e2e test suites.
3. **Allowed-to-fail** (`allowed-to-fail.yml`): runs e2e tests tagged as
   `allowed-to-fail` (currently the circular dependency tests) and posts a PR
   comment if they fail, without blocking the PR.

## Quick Reference

| Task | Command |
|---|---|
| Build everything | `bazel build //...` |
| Run unit tests | `bazel test //...` |
| Run all e2e tests | `bazel test e2e` |
| Run specific e2e suite | `bazel test e2e:<suite-name>` |
| Format Go code | `bazel run @rules_go//go -- fmt ./...` |
| Add Go dependency | `bazel run @rules_go//go -- get github.com/...@latest` |
| Sync module repos | `bazel mod tidy` |
| Format Starlark/BUILD | `bazel run //:buildifier` |
| Regenerate BUILD files | `bazel run //:gazelle` |
| Check Starlark formatting | `bazel run //:buildifier.check` |
| Check BUILD files | `bazel run //:gazelle.check` |
