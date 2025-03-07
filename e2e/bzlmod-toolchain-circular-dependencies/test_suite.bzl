"A small test suite to test the bzlmod extension"

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")

def _rpms_non_visibile_cant_be_consumed_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.expect_failure(env, "Visibility error")
    return analysistest.end(env)

rpms_non_visibile_cant_be_consumed_test = analysistest.make(
    _rpms_non_visibile_cant_be_consumed_impl,
    expect_failure = True,
)

def bazeldnf_test_suite(name):
    rpms_non_visibile_cant_be_consumed_test(
        name = "rpm_non_visibile_cant_be_consumed_test",
        target_under_test = "//:non-visibile-failure",
    )

    native.test_suite(
        name = name,
        tests = [
            ":rpm_non_visibile_cant_be_consumed_test",
        ],
    )
