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

def _rpm2tar_impl(ctx):
    rpms = []
    for rpm in ctx.files.rpms:
        rpms += ["-i", rpm.path]

    out = ctx.outputs.out
    args = ["rpm2tar", "-o", out.path] + rpms
    ctx.actions.run(
        inputs = ctx.files.rpms,
        outputs = [out],
        arguments = args,
        progress_message = "Converting %s to tar" % ctx.label.name,
        executable = ctx.executable._bazeldnf,
    )

    out = ctx.outputs.hdrs
    args = ["rpm2tar", "-o", out.path] + rpms
    ctx.actions.run(
        inputs = ctx.files.rpms,
        outputs = [out],
        arguments = args,
        progress_message = "Converting %s to tar" % ctx.label.name,
        executable = ctx.executable._bazeldnf,
    )

    out = ctx.outputs.libs
    args = ["rpm2tar", "-o", out.path] + rpms
    ctx.actions.run(
        inputs = ctx.files.rpms,
        outputs = [out],
        arguments = args,
        progress_message = "Converting %s to tar" % ctx.label.name,
        executable = ctx.executable._bazeldnf,
    )

    return [DefaultInfo(files = depset([ctx.outputs.out]))]

_rpm2tar_attrs = {
    "rpms": attr.label_list(allow_files = True),
    "_bazeldnf": attr.label(
        executable = True,
        cfg = "exec",
        allow_files = True,
        default = Label("//cmd:cmd"),
    ),
    "out": attr.output(mandatory = True),
    "libs": attr.output(mandatory = True),
    "hdrs": attr.output(mandatory = True),
}

_rpm2tar = rule(
    implementation = _rpm2tar_impl,
    attrs = _rpm2tar_attrs,
)

def rpmtree(**kwargs):
    _rpm2tar(
        out = kwargs["name"] + ".tar",
        libs = kwargs["name"] + "/libs.tar",
        hdrs = kwargs["name"] + "/hdrs.tar",
        **kwargs
    )
