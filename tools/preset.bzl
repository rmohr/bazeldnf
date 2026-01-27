"""override flags from preset.bzl

For the bazeldnf we need some special flags to not use the recommended defaults
"""
EXTRA_PRESET_FLAGS = {
    "lockfile_mode": struct(
        command = "common:ci",
        default = "update",
        description = """\
        bazeldnf CI tests against multiple bazel versions, so we can't be strict with MODULE.bazel
        """,
    ),
    "incompatible_modify_execution_info_additive": struct(
        default = True,
        if_bazel_version = False,  # hack: this flag is not supported by Bazel 6 and having it mentioned breaks bazeldnf CI
        description = "Accept multiple --modify_execution_info flags, rather than the last flag overwriting earlier ones.",
    ),
}
