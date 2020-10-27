all: gazelle buildifier

deps-update:
	bazelisk run //:gazelle -- update-repos -from_file=go.mod -prune=true

gazelle:
	bazelisk run //:gazelle

test: gazelle
	bazelisk test //pkg/...

buildifier:
	bazelisk run //:buildifier

.PHONY: gazelle test deps-update buildifier
