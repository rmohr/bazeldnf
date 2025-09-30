""" Rules generating repositories for testing. """

def _repo1_impl(rctx):
    rctx.file(
        "BUILD.bazel",
        """
load("@bazeldnf//bazeldnf:defs.bzl", "tar2files")

# This tests if `tar2files` can be called in an external repo
tar2files(
    name = "something_libs",
    files = {
        "/usr/lib64": [
            "libvirt.so.0",
            "libvirt.so.0.11000.0",
        ],
    },
    tar = "@//:something",
    visibility = ["//visibility:public"],
)
""",
    )

repo1 = repository_rule(
    implementation = _repo1_impl,
)
