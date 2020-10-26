all: gazelle

deps-update:
	bazelisk run //:gazelle -- update-repos -from_file=go.mod -prune=true

gazelle:
	bazelisk run //:gazelle

test: gazelle
	bazelisk test //pkg/...

.PHONY: gazelle
