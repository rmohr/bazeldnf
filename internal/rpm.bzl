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

load("@bazel_tools//tools/build_defs/repo:utils.bzl", "update_attrs")

_HTTP_FILE_BUILD = """
package(default_visibility = ["//visibility:public"])
filegroup(
    name = "rpm",
    srcs = ["downloaded"],
)
"""

def _rpm_impl(ctx):
    if ctx.attr.urls:
        downloaded_file_path = "downloaded"
        download_path = ctx.path("rpm/" + downloaded_file_path)
        download_info = ctx.download(
            ctx.attr.urls,
            "rpm/" + downloaded_file_path,
            ctx.attr.sha256,
        )
    else:
        fail("urls must be specified")
    ctx.file("WORKSPACE", "workspace(name = \"{name}\")".format(name = ctx.name))
    ctx.file("rpm/BUILD", _HTTP_FILE_BUILD.format(downloaded_file_path))
    return update_attrs(ctx.attr, _rpm_attrs.keys(), {"sha256": download_info.sha256})

_rpm_attrs = {
    "urls": attr.string_list(),
    "strip_prefix": attr.string(),
    "sha256": attr.string(),
}

rpm = repository_rule(
    implementation = _rpm_impl,
    attrs = _rpm_attrs,
)
