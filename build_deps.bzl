"Provides WORKSPACE build dependencies to build the bazeldnf binary"

load(
    "@bazel_gazelle//:deps.bzl",
    _go_repository = "go_repository",
)

def bazeldnf_build_dependencies():
    "dependencies to build the bazeldnf binary"
    _maybe(
        build_external = "external",
        name = "co_honnef_go_tools",
        importpath = "honnef.co/go/tools",
        sum = "h1:/hemPrYIhOhy8zYrNj+069zDB68us2sMGsfkFJO0iZs=",
        version = "v0.0.0-20190523083050-ea95bdfd59fc",
    )
    _maybe(
        build_external = "external",
        name = "com_github_bazelbuild_buildtools",
        importpath = "github.com/bazelbuild/buildtools",
        sum = "h1:DQ4RKicYtQ0KmtM6+PhBbMhd+qaazxi/8FuZoD7a1Zo=",
        version = "v0.0.0-20230127124510-cf446296fb76",
    )
    _maybe(
        build_external = "external",
        name = "com_github_burntsushi_toml",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
        version = "v0.3.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_census_instrumentation_opencensus_proto",
        importpath = "github.com/census-instrumentation/opencensus-proto",
        sum = "h1:glEXhBS5PSLLv4IXzLA5yPRVX4bilULVyxxbrfOtDAk=",
        version = "v0.2.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_chzyer_logex",
        importpath = "github.com/chzyer/logex",
        sum = "h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=",
        version = "v1.1.10",
    )
    _maybe(
        build_external = "external",
        name = "com_github_chzyer_readline",
        importpath = "github.com/chzyer/readline",
        sum = "h1:fY5BOSpyZCqRo5OhCuC+XN+r/bBCmeuuJtjz+bCNIf8=",
        version = "v0.0.0-20180603132655-2972be24d48e",
    )
    _maybe(
        build_external = "external",
        name = "com_github_chzyer_test",
        importpath = "github.com/chzyer/test",
        sum = "h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=",
        version = "v0.0.0-20180213035817-a1ea475d72b1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_client9_misspell",
        importpath = "github.com/client9/misspell",
        sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
        version = "v0.3.4",
    )
    _maybe(
        build_external = "external",
        name = "com_github_cpuguy83_go_md2man_v2",
        importpath = "github.com/cpuguy83/go-md2man/v2",
        sum = "h1:p1EgwI/C7NhT0JmVkwCD2ZBK8j4aeHQX2pMHHBfMQ6w=",
        version = "v2.0.2",
    )
    _maybe(
        build_external = "external",
        name = "com_github_crillab_gophersat",
        importpath = "github.com/crillab/gophersat",
        sum = "h1:l4fgnEMmy1+b7pn3nvPwj1ja3Z9MgXE4hUIl9TU8v+M=",
        version = "v1.3.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_envoyproxy_go_control_plane",
        importpath = "github.com/envoyproxy/go-control-plane",
        sum = "h1:4cmBvAEBNJaGARUEs3/suWRyfyBfhf7I60WBZq+bv2w=",
        version = "v0.9.1-0.20191026205805-5f8ba28d4473",
    )
    _maybe(
        build_external = "external",
        name = "com_github_envoyproxy_protoc_gen_validate",
        importpath = "github.com/envoyproxy/protoc-gen-validate",
        sum = "h1:EQciDnbrYxy13PgWoY8AqoxGiPrpgBZ1R8UNe3ddc+A=",
        version = "v0.1.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_go_logr_logr",
        importpath = "github.com/go-logr/logr",
        sum = "h1:2DntVwHkVopvECVRSlL5PSo9eG+cAkDCuckLubN+rq0=",
        version = "v1.2.3",
    )
    _maybe(
        build_external = "external",
        name = "com_github_golang_glog",
        importpath = "github.com/golang/glog",
        sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
        version = "v0.0.0-20160126235308-23def4e6c14b",
    )
    _maybe(
        build_external = "external",
        name = "com_github_golang_mock",
        importpath = "github.com/golang/mock",
        sum = "h1:G5FRp8JnTd7RQH5kemVNlMeyXQAztQ3mOWV95KxsXH8=",
        version = "v1.1.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_golang_protobuf",
        importpath = "github.com/golang/protobuf",
        sum = "h1:ROPKBNFfQgOUMifHyP+KYbvpjbdoFNs+aK7DXlji0Tw=",
        version = "v1.5.2",
    )
    _maybe(
        build_external = "external",
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        sum = "h1:O2Tfq5qg4qc4AmwVlvv0oLiVAGB7enBSJ2x2DqQFi38=",
        version = "v0.5.9",
    )
    _maybe(
        build_external = "external",
        name = "com_github_inconshreveable_mousetrap",
        importpath = "github.com/inconshreveable/mousetrap",
        sum = "h1:U3uMjPSQEBMNp1lFxmllqCPM6P5u/Xq7Pgzkat/bFNc=",
        version = "v1.0.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_klauspost_compress",
        importpath = "github.com/klauspost/compress",
        sum = "h1:bPb7nMRdOZYDrpPMTA3EInUQrdgoBinqUuSwlGdKDdE=",
        version = "v1.11.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_onsi_ginkgo_v2",
        importpath = "github.com/onsi/ginkgo/v2",
        sum = "h1:/XxtEV3I3Eif/HobnVx9YmJgk8ENdRsuUmM+fLCFNow=",
        version = "v2.7.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_onsi_gomega",
        importpath = "github.com/onsi/gomega",
        sum = "h1:03cDLK28U6hWvCAns6NeydX3zIm4SF3ci69ulidS32Q=",
        version = "v1.26.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_prometheus_client_model",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:gQz4mCbXsO+nc9n1hCxHcGA3Zx3Eo+UHZoInFGUIXNM=",
        version = "v0.0.0-20190812154241-14fe0d1b01d4",
    )
    _maybe(
        build_external = "external",
        name = "com_github_russross_blackfriday_v2",
        importpath = "github.com/russross/blackfriday/v2",
        sum = "h1:JIOH55/0cWyOuilr9/qlrm0BSXldqnqwMsf35Ld67mk=",
        version = "v2.1.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_sassoftware_go_rpmutils",
        importpath = "github.com/sassoftware/go-rpmutils",
        replace = "github.com/rmohr/go-rpmutils",
        sum = "h1:IPm+rJWi7P9mkoFQjloDEGThxlM71dzDrC2DZSjgnL8=",
        version = "v0.1.2-0.20201215123907-5acf7436c00d",
    )
    _maybe(
        build_external = "external",
        name = "com_github_sirupsen_logrus",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:trlNQbNUG3OdDrDil03MCb1H2o9nJ1x4/5LYw7byDE0=",
        version = "v1.9.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_spf13_cobra",
        importpath = "github.com/spf13/cobra",
        sum = "h1:o94oiPyS4KD1mPy2fmcYYHHfCxLqYjJOhGsCHFZtEzA=",
        version = "v1.6.1",
    )
    _maybe(
        build_external = "external",
        name = "com_github_spf13_pflag",
        importpath = "github.com/spf13/pflag",
        sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
        version = "v1.0.5",
    )
    _maybe(
        build_external = "external",
        name = "com_github_stretchr_objx",
        importpath = "github.com/stretchr/objx",
        sum = "h1:4G4v2dO3VZwixGIRoQ5Lfboy6nUhCyYzaqnIAPPhYs4=",
        version = "v0.1.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_stretchr_testify",
        importpath = "github.com/stretchr/testify",
        sum = "h1:nwc3DEeHmmLAfoZucVR881uASk0Mfjw8xYJ99tb5CcY=",
        version = "v1.7.0",
    )
    _maybe(
        build_external = "external",
        name = "com_github_xi2_xz",
        importpath = "github.com/xi2/xz",
        sum = "h1:nIPpBwaJSVYIxUFsDv3M8ofmx9yWTog9BfvIu0q41lo=",
        version = "v0.0.0-20171230120015-48954b6210f8",
    )
    _maybe(
        build_external = "external",
        name = "com_google_cloud_go",
        importpath = "cloud.google.com/go",
        sum = "h1:e0WKqKTd5BnrG8aKH3J3h+QvEIQtSUcf2n5UZ5ZgLtQ=",
        version = "v0.26.0",
    )
    _maybe(
        build_external = "external",
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:yhCVgyC4o1eVCa2tZl7eS0r+SDo693bJlVdllGtEeKM=",
        version = "v0.0.0-20161208181325-20d25e280405",
    )
    _maybe(
        build_external = "external",
        name = "in_gopkg_yaml_v2",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
        version = "v2.4.0",
    )
    _maybe(
        build_external = "external",
        name = "in_gopkg_yaml_v3",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
        version = "v3.0.1",
    )
    _maybe(
        build_external = "external",
        name = "io_k8s_sigs_yaml",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:a2VclLzOGrwOHDiV8EfBGhvjHvP46CtW5j6POvhYGGo=",
        version = "v1.3.0",
    )
    _maybe(
        build_external = "external",
        name = "net_starlark_go",
        importpath = "go.starlark.net",
        sum = "h1:xwwDQW5We85NaTk2APgoN9202w/l0DVGp+GZMfsrh7s=",
        version = "v0.0.0-20210223155950-e043a3d3c984",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_google_appengine",
        importpath = "google.golang.org/appengine",
        sum = "h1:/wp5JvzpHIxhs/dumFmF7BXTf3Z+dd4uXta4kVyO508=",
        version = "v1.4.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_google_genproto",
        importpath = "google.golang.org/genproto",
        sum = "h1:+kGHl1aib/qcwaRi1CbqBZ1rk19r85MNUf8HaBghugY=",
        version = "v0.0.0-20200526211855-cb27e3aa2013",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_google_grpc",
        importpath = "google.golang.org/grpc",
        sum = "h1:rRYRFMVgRv6E0D70Skyfsr28tDXIuuPZyWGMPdMcnXg=",
        version = "v1.27.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_google_protobuf",
        importpath = "google.golang.org/protobuf",
        sum = "h1:w43yiav+6bVFTBQFZX0r7ipe9JQ1QsbMgHwbBziscLw=",
        version = "v1.28.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_crypto",
        importpath = "golang.org/x/crypto",
        sum = "h1:vEg9joUBmeBcK9iSJftGNf3coIG4HqZElCPehJsfAYM=",
        version = "v0.0.0-20200604202706-70a84ac30bf9",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_exp",
        importpath = "golang.org/x/exp",
        sum = "h1:c2HOrn5iMezYjSlGPncknSEr/8x5LELb/ilJbXi9DEA=",
        version = "v0.0.0-20190121172915-509febef88a4",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_lint",
        importpath = "golang.org/x/lint",
        sum = "h1:XQyxROzUlZH+WIQwySDgnISgOivlhjIEwaQaJEJrrN0=",
        version = "v0.0.0-20190313153728-d0100b6bd8b3",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_mod",
        importpath = "golang.org/x/mod",
        sum = "h1:6zppjxzCulZykYSLyVDYbneBfbaBIQPYMevg0bEwv2s=",
        version = "v0.6.0-dev.0.20220419223038-86c51ed26bb4",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_net",
        importpath = "golang.org/x/net",
        sum = "h1:GyT4nK/YDHSqa1c4753ouYCDajOYKTja9Xb/OHtgvSw=",
        version = "v0.5.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_oauth2",
        importpath = "golang.org/x/oauth2",
        sum = "h1:vEDujvNQGv4jgYKudGeI/+DAX4Jffq6hpD55MmoEvKs=",
        version = "v0.0.0-20180821212333-d2e6202438be",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_sync",
        importpath = "golang.org/x/sync",
        sum = "h1:8gQV6CLnAEikrhgkHFbMAEhagSSnXWGV915qUMm9mrU=",
        version = "v0.0.0-20190423024810-112230192c58",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:Zr2JFtRQNX3BCZ8YtxRE9hNJYC8J6I1MVbMg6owUp18=",
        version = "v0.4.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_term",
        importpath = "golang.org/x/term",
        sum = "h1:O7UWfv5+A2qiuulQk30kVinPoMtoIPeVaKLEgLpVkvg=",
        version = "v0.4.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_text",
        importpath = "golang.org/x/text",
        sum = "h1:3XmdazWV+ubf7QgHSTWeykHOci5oeekaGJBLkrkaw4k=",
        version = "v0.6.0",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:VveCTK38A2rkS8ZqFY25HIDFscX5X9OoEhJd3quQmXU=",
        version = "v0.1.12",
    )
    _maybe(
        build_external = "external",
        name = "org_golang_x_xerrors",
        importpath = "golang.org/x/xerrors",
        sum = "h1:go1bK/D/BFZV2I8cIQd1NKEZ+0owSTG1fDTci4IqFcE=",
        version = "v0.0.0-20200804184101-5ec99f83aff1",
    )

def _maybe(name, **kwargs):
    if name not in native.existing_rules():
        _go_repository(name = name, **kwargs)
