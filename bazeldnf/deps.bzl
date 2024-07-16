"""bazeldnf public dependency for WORKSPACE"""

load(
    "@bazeldnf//bazeldnf:repositories.bzl",
    _bazeldnf_dependencies = "bazeldnf_dependencies",
    _bazeldnf_register_toolchains = "bazeldnf_register_toolchains",
)

def bazeldnf_dependencies(name = "bazeldnf", **kwargs):
    """bazeldnf dependencies when consuming bazeldnf through WORKSPACE"""
    _bazeldnf_dependencies()
    _bazeldnf_register_toolchains(name = name, **kwargs)
