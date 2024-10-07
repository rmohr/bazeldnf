"bazeldnf repo integration test dependencies"

load("@bazeldnf//bazeldnf:defs.bzl", "rpm")

def bazeldnf_test_dependencies():
    rpm(
        name = "libvirt-libs-6.1.0-2.fc32.x86_64.rpm",
        sha256 = "3a0a3d88c6cb90008fbe49fe05e7025056fb9fa3a887c4a78f79e63f8745c845",
        urls = [
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/32/Everything/x86_64/os/Packages/l/libvirt-libs-6.1.0-2.fc32.x86_64.rpm",
            "https://storage.googleapis.com/builddeps/3a0a3d88c6cb90008fbe49fe05e7025056fb9fa3a887c4a78f79e63f8745c845",
        ],
    )

    rpm(
        name = "libvirt-devel-6.1.0-2.fc32.x86_64.rpm",
        sha256 = "2ebb715341b57a74759aff415e0ff53df528c49abaa7ba5b794b4047461fa8d6",
        urls = [
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/32/Everything/x86_64/os/Packages/l/libvirt-devel-6.1.0-2.fc32.x86_64.rpm",
            "https://storage.googleapis.com/builddeps/2ebb715341b57a74759aff415e0ff53df528c49abaa7ba5b794b4047461fa8d6",
        ],
    )
