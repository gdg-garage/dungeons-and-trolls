load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_proto_grpc",
    sha256 = "928e4205f701b7798ce32f3d2171c1918b363e9a600390a25c876f075f1efc0a",
    strip_prefix = "rules_proto_grpc-4.4.0",
    urls = ["https://github.com/rules-proto-grpc/rules_proto_grpc/releases/download/4.4.0/rules_proto_grpc-4.4.0.tar.gz"],
)

load("@rules_proto_grpc//:repositories.bzl", "rules_proto_grpc_toolchains", "rules_proto_grpc_repos")
rules_proto_grpc_toolchains()
rules_proto_grpc_repos()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")
rules_proto_dependencies()
rules_proto_toolchains()

load("@rules_proto_grpc//:repositories.bzl", "bazel_gazelle", "io_bazel_rules_go")  # buildifier: disable=same-origin-load
io_bazel_rules_go()
bazel_gazelle()

load("@rules_proto_grpc//go:repositories.bzl", rules_proto_grpc_go_repos = "go_repos")
rules_proto_grpc_go_repos()

load("@rules_proto_grpc//grpc-gateway:repositories.bzl", rules_proto_grpc_gateway_repos = "gateway_repos")
rules_proto_grpc_gateway_repos()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
go_rules_dependencies()

go_register_toolchains(
    version = "1.19",
)

load("@com_github_grpc_ecosystem_grpc_gateway_v2//:repositories.bzl", "go_repositories")
go_repositories()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()

# grpc_gateway_version = "2.5.0"
# grpc_gateway_sha = "e8334c270a479f55ad9f264e798680ac536f473d7711593f6eadab3df2d1ddc3"
# http_archive(
#     name = "com_github_grpc_ecosystem_grpc_gateway",
#     sha256 = grpc_gateway_sha,
#     strip_prefix = "grpc-gateway-%s" % grpc_gateway_version,
#     urls = ["https://github.com/grpc-ecosystem/grpc-gateway/archive/refs/tags/v%s.zip" % grpc_gateway_version],
# )

# load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# http_archive(
#     name = "com_google_protobuf",
#     sha256 = "d0f5f605d0d656007ce6c8b5a82df3037e1d8fe8b121ed42e536f569dec16113",
#     strip_prefix = "protobuf-3.14.0",
#     urls = [
#         "https://mirror.bazel.build/github.com/protocolbuffers/protobuf/archive/v3.14.0.tar.gz",
#         "https://github.com/protocolbuffers/protobuf/archive/v3.14.0.tar.gz",
#     ],
# )

# load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

# protobuf_deps()


# http_archive(
#     name = "io_bazel_rules_go",
#     sha256 = "6b65cb7917b4d1709f9410ffe00ecf3e160edf674b78c54a894471320862184f",
#     urls = [
#         "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.39.0/rules_go-v0.39.0.zip",
#         "https://github.com/bazelbuild/rules_go/releases/download/v0.39.0/rules_go-v0.39.0.zip",
#     ],
# )

# load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

# go_rules_dependencies()

# go_register_toolchains(version = "1.19.3")

# TODO this is old

# load("@org_pubref_rules_protobuf//grpc_gateway:rules.bzl", "grpc_gateway_proto_repositories")
# grpc_gateway_proto_repositories()


# NEW

http_archive(
    name = "build_stack_rules_proto",
    sha256 = "ac7e2966a78660e83e1ba84a06db6eda9a7659a841b6a7fd93028cd8757afbfb",
    strip_prefix = "rules_proto-2.0.1",
    urls = ["https://github.com/stackb/rules_proto/archive/v2.0.1.tar.gz"],
)

register_toolchains("@build_stack_rules_proto//toolchain:standard")