load(
    "@bazeldnf//bazeldnf:defs.bzl",
    _bazeldnf = "bazeldnf",
)

def bazeldnf(*args, **kwargs):
    print("import this method from @bazeldnf//bazeldnf:defs.bzl")
    _bazeldnf(*args, **kwargs)
