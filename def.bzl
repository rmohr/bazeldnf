"legacy API"

load(
    "@bazeldnf//bazeldnf:defs.bzl",
    _bazeldnf = "bazeldnf",
)

def bazeldnf(*args, **kwargs):
    # buildifier: disable=print
    print("import this method from @bazeldnf//bazeldnf:defs.bzl")
    _bazeldnf(*args, **kwargs)
