"A small test suite to test the bzlmod extension"

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")

def _rpms_listed_only_once_impl(ctx):
    env = analysistest.begin(ctx)

    actions = analysistest.target_actions(env)
    asserts.equals(env, 1, len(actions))
    inputs = actions[0].inputs.to_list()
    all_rpms = dict()

    for f in inputs:
        if not f.basename.endswith(".rpm"):
            continue
        asserts.equals(env, False, f.basename in all_rpms)
        all_rpms[f.basename] = 1

    return analysistest.end(env)

rpms_listed_only_once_test = analysistest.make(_rpms_listed_only_once_impl)

def _rpms_non_visibile_cant_be_consumed_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.expect_failure(env, "Visibility error")
    return analysistest.end(env)

rpms_non_visibile_cant_be_consumed_test = analysistest.make(
    _rpms_non_visibile_cant_be_consumed_impl,
    expect_failure = True,
)

def bazeldnf_test_suite(name):
    rpms_listed_only_once_test(
        name = "rpm_listed_only_once_test",
        target_under_test = "//:something",
    )

    rpms_non_visibile_cant_be_consumed_test(
        name = "rpm_non_visibile_cant_be_consumed_test",
        target_under_test = "//:non-visibile-failure",
    )

    native.test_suite(
        name = name,
        tests = [
            ":rpm_listed_only_once_test",
            ":rpm_non_visibile_cant_be_consumed_test",
        ],
    )
