"Platforms definition for this repository"

# Add more platforms as needed to mirror all the binaries
# published by the upstream project.
PLATFORMS = {
    "darwin-amd64": struct(
        compatible_with = [
            "@platforms//os:osx",
            "@platforms//cpu:x86_64",
        ],
    ),
    "darwin-arm64": struct(
        compatible_with = [
            "@platforms//os:osx",
            "@platforms//cpu:arm64",
        ],
    ),
    "linux-amd64": struct(
        compatible_with = [
            "@platforms//os:linux",
            "@platforms//cpu:x86_64",
        ],
    ),
    "linux-arm64": struct(
        compatible_with = [
            "@platforms//os:linux",
            "@platforms//cpu:arm64",
        ],
    ),
    "linux-ppc64": struct(
        compatible_with = [
            "@platforms//os:linux",
            "@platforms//cpu:ppc",
        ],
    ),
    "linux-ppc64le": struct(
        compatible_with = [
            "@platforms//os:linux",
            "@platforms//cpu:ppc",
        ],
    ),
    "linux-s390x": struct(
        compatible_with = [
            "@platforms//os:linux",
            "@platforms//cpu:s390x",
        ],
    ),
}
