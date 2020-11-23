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

    return [DefaultInfo(files = depset([ctx.outputs.out]))]

def _tar2files_impl(ctx):
    out = ctx.outputs.out
    files = []
    for out in ctx.outputs.out:
        files += [out.path]

    args = ["tar2files", "--file-prefix", ctx.attr.prefix, "-i", ctx.files.tar[0].path] + files
    ctx.actions.run(
        inputs = ctx.files.tar,
        outputs = ctx.outputs.out,
        arguments = args,
        progress_message = "Extracting files",
        executable = ctx.executable._bazeldnf,
    )

    return [DefaultInfo(files = depset(ctx.outputs.out))]

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

_tar2files_attrs = {
    "tar": attr.label(allow_single_file = True),
    "_bazeldnf": attr.label(
        executable = True,
        cfg = "exec",
        allow_files = True,
        default = Label("//cmd:cmd"),
    ),
    "prefix": attr.string(),
    "out": attr.output_list(mandatory = True),
}

_rpm2tar = rule(
    implementation = _rpm2tar_impl,
    attrs = _rpm2tar_attrs,
)

_tar2files = rule(
    implementation = _tar2files_impl,
    attrs = _tar2files_attrs,
)

def rpmtree(**kwargs):
    files = kwargs["files"]
    kwargs.pop("files", None)
    basename = kwargs["name"]
    kwargs.pop("name", None)
    tarname = basename + ".tar"
    _rpm2tar(
        name = basename,
        out = tarname,
        **kwargs
    )
    kwargs.pop("rpms", None)
    if files:
        for k, v in files.items():
            name = basename + k
            files = []
            for file in v:
                files = files + [basename + "/" + file]
            _tar2files(
                name = name,
                prefix = k,
                tar = tarname,
                out = files,
                **kwargs
            )
