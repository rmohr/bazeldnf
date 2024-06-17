"Make releases for platforms supported by bazeldnf"

load("@aspect_bazel_lib//lib:copy_file.bzl", "copy_file")
load("@aspect_bazel_lib//lib:transitions.bzl", "platform_transition_filegroup")
load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")
load("@aspect_bazel_lib//tools/release:hashes.bzl", "hashes")
load("@bazeldnf//bazeldnf/private:toolchains_repo.bzl", "PLATFORMS")

# buildozer: disable=function-docstring
def build_for_platform(name, value):

    # define a target platform first because we may not have one
    native.platform(
        name = name,
        constraint_values = value.compatible_with,
        visibility = ["//visibility:public"],
    )

    # create a binary for the target platform
    build = "bazeldnf_{}_build".format(name)
    platform_transition_filegroup(
        name = build,
        srcs = ["@bazeldnf//cmd"],
        target_platform = ":{}".format(name),
    )

    artifact = "bazeldnf_{}".format(name)
    copy_file(
        name = "copy_{}".format(build),
        src = build,
        out = artifact,
    )

    # compuate the sha256 of the binary
    hashes(
        name = "bazeldnf_{}.sha256".format(name),
        src = artifact,
    )

    return [artifact, "bazeldnf_{}.sha256".format(name)]

# buildozer: disable=function-docstring
def build_for_all_platforms(name, **kwargs):
    outs = []

    for k, v in PLATFORMS.items():
        outs.extend(build_for_platform(name = k, value = v))

    write_source_files(
        name = name,
        files = dict([ ["latest/{}".format(x), ":{0}".format(x)] for x in outs ]),
        diff_test = False,
        check_that_out_file_exists = False,
        **kwargs,
    )
