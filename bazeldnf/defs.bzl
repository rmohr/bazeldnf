"""
Public API to use bazeldnf from other repositories
"""

load(
    "@bazeldnf//internal:bazeldnf.bzl",
    _bazeldnf = "bazeldnf",
)
load(
    "@bazeldnf//internal:rpm.bzl",
    _rpm = "rpm",
)
load(
    "@bazeldnf//internal:rpmtree.bzl",
    _rpmtree = "rpmtree",
    _tar2files = "tar2files",
)
load(
    "@bazeldnf//internal:xattrs.bzl",
    _xattrs = "xattrs",
)

bazeldnf = _bazeldnf
rpm = _rpm
rpmtree = _rpmtree
tar2files = _tar2files
xattrs = _xattrs
