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

"Modify xattrs on tar file members"

load("//bazeldnf:toolchain.bzl", "BAZELDNF_TOOLCHAIN")

def _xattrs_impl(ctx):
    out = ctx.outputs.out
    args = ["xattr", "--input", ctx.files.tar[0].path, "--output", out.path]

    if ctx.attr.capabilities:
        capabilities = []
        for k, v in ctx.attr.capabilities.items():
            capabilities.append(k + "=" + ":".join(v))
        args += ["--capabilities", ",".join(capabilities)]

    if ctx.attr.selinux_labels:
        selinux_labels = []
        for k, v in ctx.attr.selinux_labels.items():
            selinux_labels.append([k + "=" + v])
        args += ["--selinux-labels", ",".join(selinux_labels)]

    bazeldnf = ctx.toolchains[BAZELDNF_TOOLCHAIN]._tool

    ctx.actions.run(
        inputs = ctx.files.tar,
        outputs = [out],
        arguments = args,
        progress_message = "Enriching %s with xattrs" % ctx.label.name,
        executable = bazeldnf,
    )

    return [DefaultInfo(files = depset([ctx.outputs.out]))]

_xattrs_attrs = {
    "tar": attr.label(allow_single_file = True),
    "capabilities": attr.string_list_dict(),
    "selinux_labels": attr.string_dict(),
    "out": attr.output(mandatory = True),
}

_xattrs = rule(
    implementation = _xattrs_impl,
    attrs = _xattrs_attrs,
    toolchains = [BAZELDNF_TOOLCHAIN],
)

def xattrs(**kwargs):
    basename = kwargs["name"]
    tarname = basename + ".tar"
    _xattrs(
        out = tarname,
        **kwargs
    )
