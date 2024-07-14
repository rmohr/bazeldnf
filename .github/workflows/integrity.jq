# JQ filter to transform sha256 files to a value we can read from starlark.
# NB: the sha256 files are expected to be newline-terminated.
#
# Input looks like
# 48552e399a1f2ab97e62ca7fce5783b6214e284330c7555383f43acf82446636 bazeldnf-v0.6.0-rc7-linux-amd64\n...
#
# Output should look like
# {
#  "linux-amd64": "48552e399a1f2ab97e62ca7fce5783b6214e284330c7555383f43acf82446636",
#  ...
# }

.
| sub($ARGS.named.PREFIX; ""; "g")
| rtrimstr("\n")
| split("\n")
| map(split(" "))
| map({"key": .[1], "value": .[0]})
| from_entries
