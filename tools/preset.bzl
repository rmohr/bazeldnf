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
    #"incompatible_modify_execution_info_additive": struct(
    #    default = True,
    #    if_bazel_version = False,  # hack: this flag is not supported by Bazel 6 and having it mentioned breaks bazeldnf CI
    #    description = "Accept multiple --modify_execution_info flags, rather than the last flag overwriting earlier ones.",
    #),
    "incompatible_enforce_starlark_utf8": struct(
        default = False,
        if_bazel_version = False,  # hack: this flag is not supported with bazel 7
        description = "ignored flag"
    ),
    "incompatible_enable_proto_toolchain_resolution": struct(
        default = True,
        description = """\
        Bazel 7 introduced this flag to allow us fetch `protoc` rather than re-build it!
        That flag ALSO decouples how each built-in language rule (Java, Python, C++, etc.) locates the runtime.
        """,
    ),
    "module_mirrors": struct(
        default = False,
        if_bazel_version = False,  # hack: this flag is not supported with bazel 7
        description = "ignroed flag"
    ),
    "per_file_copt": [
        struct(
            default = "external/.*protobuf.*@--PROTOBUF_WAS_NOT_SUPPOSED_TO_BE_BUILT",
            description = "Make sure protobuf is not built from source",
        ),
        struct(
            default = "external/.*grpc.*@--GRPC_WAS_NOT_SUPPOSED_TO_BE_BUILT",
            description = "Make sure grpc is not built from source",
            allow_repeated = True,
        ),
    ],
    "host_per_file_copt": [
        struct(
            default = "external/.*protobuf.*@--PROTOBUF_WAS_NOT_SUPPOSED_TO_BE_BUILT",
            description = "Make sure protobuf is not built from source",
        ),
        struct(
            default = "external/.*grpc.*@--GRPC_WAS_NOT_SUPPOSED_TO_BE_BUILT",
            description = "Make sure grpc is not built from source",
            allow_repeated = True,
        ),
    ],
    "@protobuf//bazel/toolchains:prefer_prebuilt_protoc": struct(
        command = "common",
        default = True,
        description = "Make sure we use prebuilt protoc in bazel 7.x and 8.x"
    )
}
