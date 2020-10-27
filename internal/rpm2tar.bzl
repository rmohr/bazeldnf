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
    args = ["rpm2tar", "-o", ctx.outputs.out.path]
    for rpm in ctx.files.rpms:
        args += ["-i", rpm.path]
    ctx.actions.run(
        inputs = ctx.files.rpms,
        outputs = [ctx.outputs.out],
        arguments = args,
        progress_message = "Converting %s to tar" % ctx.label.name,
        executable = ctx.executable._bazeldnf,
    )

_rpm2tar_attrs = {
    "rpms": attr.label_list(allow_files = True),
    "_bazeldnf": attr.label(
        executable = True,
        cfg = "exec",
        allow_files = True,
        default = Label("//cmd:cmd"),
    ),
    "out": attr.output(mandatory = True),
}

_rpm2tar = rule(
    implementation = _rpm2tar_impl,
    attrs = _rpm2tar_attrs,
)

def rpm2tar(**kwargs):
    _rpm2tar(
        out = kwargs["name"] + ".tar",
        **kwargs
    )
