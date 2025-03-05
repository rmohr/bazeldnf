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

"Exposes rpm files to a Bazel workspace"

load("@bazel_tools//tools/build_defs/repo:utils.bzl", "update_attrs")

RpmInfo = provider(
    doc = """\
        Information about an RPM file

        Used to pass information about an RPM file from the rpm rule to
        other rules like rpmtree
    """,
    fields = {
        "deps": "depset of other dependencies",
        "file": "label of the RPM file",
    },
)

def _rpm_rule_impl(ctx):
    """\
    Implementation for the rpm rule

    Allows to pass information about an RPM file to other rules
    like rpmtree, keeping track of the dependency tree
    """
    deps_list = []

    for dep in ctx.attr.deps:
        deps_list.append(dep[RpmInfo].deps)

    rpm_info = RpmInfo(
        file = ctx.file.file,
        deps = depset(direct = [ctx.file.file], transitive = deps_list),
    )

    return [
        rpm_info,
        DefaultInfo(
            files = depset(direct = [ctx.file.file], transitive = deps_list),
        ),
    ]

rpm_rule = rule(
    implementation = _rpm_rule_impl,
    attrs = {
        "file": attr.label(allow_single_file = True, mandatory = True),
        "deps": attr.label_list(providers = [RpmInfo]),
    },
)

_HTTP_FILE_BUILD = """
load("@bazeldnf//internal:rpm.bzl", "rpm_rule")
package(default_visibility = ["//visibility:public"])
rpm_rule(
    name = "rpm",
    file = "{downloaded_file_path}",
    deps = [{deps}],
)
"""

def _rpm_impl(ctx):
    if ctx.attr.urls:
        downloaded_file_path = "downloaded"
        download_info = ctx.download(
            url = ctx.attr.urls,
            output = "rpm/" + downloaded_file_path,
            sha256 = ctx.attr.sha256,
            integrity = ctx.attr.integrity,
        )
    else:
        fail("urls must be specified")
    ctx.file("WORKSPACE", "workspace(name = \"{name}\")".format(name = ctx.name))
    ctx.file(
        "rpm/BUILD",
        _HTTP_FILE_BUILD.format(
            downloaded_file_path = downloaded_file_path,
            deps = ", ".join(["\"%s\"" % dep for dep in ctx.attr.dependencies]),
        ),
    )
    return update_attrs(ctx.attr, _rpm_attrs.keys(), {"integrity": download_info.integrity})

_rpm_attrs = {
    "urls": attr.string_list(),
    "sha256": attr.string(),
    "integrity": attr.string(),
    "dependencies": attr.label_list(
        mandatory = False,
        providers = [RpmInfo],
    ),
}

rpm = repository_rule(
    implementation = _rpm_impl,
    attrs = _rpm_attrs,
)
