""" Strategies for alias generation in alias repository """

load("@bazeldnf//internal:rpm.bzl", "null_rpm_rule")

def default(name, rpms, visibility = ["//visibility:public"]):
    """
    Default behaviour for alias generation.

    Everything depends on how many times was the given package resolved ("installed"):
     0 – it was requested, but not resolved – return empty providers
     1 – resolved – package available under its name
    >1 – resolved in multiple architectures (not supported at the moment)

    Args:
      name: default target name
      rpms: list of RPM metadata; each one is a dict consisting of:
        package (optional), id, repo_name, arch (optional)
        Consult `bazeldnf/extension.bzl`'s `packages_metadata` variable for more datails.
      visibility: visibility for aliases
    """

    def alias(name, rpm):
        native.alias(
            name = name,
            actual = "@{}//rpm".format(rpm["repo_name"]),
            visibility = visibility,
        )

    for rpm in rpms:
        if "arch" not in rpm or "package" not in rpm:
            continue
        alias(
            name = rpm["package"] + "." + rpm["arch"],
            rpm = rpm,
        )

    if len(rpms) == 1:
        rpm = rpms[0]
        alias(
            name = name,
            rpm = rpm,
        )

    if len(rpms) == 0:
        null_rpm_rule(
            name = name,
            visibility = visibility,
        )
