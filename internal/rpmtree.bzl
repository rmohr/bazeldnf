# Copyright 2014 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Provide helpers to convert rpm files into a single tar file

This file exposes rpmtree and tar2files to convert a group of
rpm files into either a .tar or extract files from that tar to
make available to bazel
"""

load("@bazel_skylib//lib:paths.bzl", "paths")
load("//bazeldnf:toolchain.bzl", "BAZELDNF_TOOLCHAIN")
load("//internal:rpm.bzl", "RpmInfo")

def _rpm2tar_impl(ctx):
    args = ctx.actions.args()

    out = ctx.outputs.out
    args.add_all(["rpm2tar", "--output", out])

    if ctx.attr.symlinks:
        symlinks = []
        for k, v in ctx.attr.symlinks.items():
            symlinks.append(k + "=" + v)
        args.add_joined("--symlinks", symlinks, join_with = ",")

    if ctx.attr.capabilities:
        capabilities = []
        for k, v in ctx.attr.capabilities.items():
            capabilities.append(k + "=" + ":".join(v))
        args.add_joined("--capabilities", capabilities, join_with = ",")

    if ctx.attr.selinux_labels:
        selinux_labels = []
        for k, v in ctx.attr.selinux_labels.items():
            selinux_labels.append(k + "=" + v)
        args.add_joined("--selinux-labels", selinux_labels, join_with = ",")

    all_rpms = []

    for target in ctx.attr.rpms:
        if target[RpmInfo].file not in all_rpms:
            all_rpms.append(target[RpmInfo].file)

        for rpm in target[RpmInfo].deps:
            if rpm not in all_rpms:
                all_rpms.append(rpm)

    for rpm in all_rpms:
        args.add_all(["--input", rpm.path])

    ctx.actions.run(
        inputs = ctx.files.rpms,
        outputs = [out],
        arguments = [args],
        mnemonic = "Rpm2Tar",
        progress_message = "Converting %s to tar" % ctx.label.name,
        executable = ctx.toolchains[BAZELDNF_TOOLCHAIN]._tool,
    )

    return [DefaultInfo(files = depset([ctx.outputs.out]))]

def _expand_path(files):
    return [x.path for x in files]

def _tar2files_impl(ctx):
    args = ctx.actions.args()
    strip_prefix = paths.join(
        ctx.bin_dir.path,
        ctx.label.package,
        ctx.label.name,
    )

    args.set_param_file_format("multiline")
    args.use_param_file("@%s")
    args.add_all([
        "tar2files",
        "--file-prefix",
        ctx.attr.prefix,
        "--strip-prefix",
        strip_prefix,
        "--input",
        ctx.files.tar[0],
    ])
    args.add_all([ctx.outputs.out], map_each = _expand_path)

    ctx.actions.run(
        inputs = ctx.files.tar,
        outputs = ctx.outputs.out,
        arguments = [args],
        mnemonic = "Tar2Files",
        progress_message = "Extracting files",
        executable = ctx.toolchains[BAZELDNF_TOOLCHAIN]._tool,
    )

    return [DefaultInfo(files = depset(ctx.outputs.out))]

_rpm2tar_attrs = {
    "rpms": attr.label_list(allow_files = True, providers = [RpmInfo]),
    "symlinks": attr.string_dict(),
    "capabilities": attr.string_list_dict(),
    "selinux_labels": attr.string_list_dict(),
    "out": attr.output(mandatory = True),
}

_tar2files_attrs = {
    "tar": attr.label(allow_single_file = True),
    "prefix": attr.string(),
    "out": attr.output_list(mandatory = True),
}

_rpm2tar = rule(
    implementation = _rpm2tar_impl,
    attrs = _rpm2tar_attrs,
    toolchains = [BAZELDNF_TOOLCHAIN],
)

_tar2files = rule(
    implementation = _tar2files_impl,
    attrs = _tar2files_attrs,
    toolchains = [BAZELDNF_TOOLCHAIN],
)

def rpmtree(name, **kwargs):
    """Creates a tar file from a list of rpm files."""
    tarname = name + ".tar"
    _rpm2tar(
        name = name,
        out = tarname,
        **kwargs
    )

def tar2files(name, files = None, **kwargs):
    """Extracts files from a tar file.

    Args:
        name: The name of the tar file to be processed.
        files: A dictionary where each key-value pair represents a file to be extracted.
               If not provided, the function will fail.
        **kwargs: Additional keyword arguments to be passed to the _tar2files function.
    """
    if not files:
        fail("files is a required attribute")

    basename = name
    for k, v in files.items():
        name = basename + k
        files = []
        for file in v:
            files.append(name + "/" + file)
        _tar2files(
            name = name,
            prefix = k,
            out = files,
            **kwargs
        )
