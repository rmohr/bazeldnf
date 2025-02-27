all: test

test:
	bazelisk run @bazeldnf_rpms//:fetch-repo
	bazelisk run @bazeldnf_rpms//:update-lock-file
	bazelisk build //...
