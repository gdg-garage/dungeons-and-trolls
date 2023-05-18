# load("@rules_proto_grpc//go:defs.bzl", "go_proto_compile")
load("@rules_proto_grpc//go:defs.bzl", "go_grpc_compile")
load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_grpc_compile")

# go_proto_compile(
#     name = "dungeonsandtrolls_go_proto",
#     protos = ["//proto:dungeonsandtrolls_proto"],
#     output_mode = "NO_PREFIX"S
# )


## gazelle:proto_strip_import_prefix proto
## xxxgazelle:go_grpc_compilers	@io_bazel_rules_go//proto:go_grpc
## gazelle:go_grpc_compilers	?

# go_grpc_compile(
#     name = "dungeonsandtrolls",
#     protos = ["//proto:dungeonsandtrolls_proto"],
#     # compilers = ["@io_bazel_rules_go//proto:go_grpc"],
#     # output_mode = "NO_PREFIX",
#     # prefix_path = "xxx",
# )

gateway_grpc_compile(
    name = "dungeonsandtrolls_gw",
    protos = ["//proto:dungeonsandtrolls_proto"],
    output_mode = "NO_PREFIX",
    prefix_path = "dungeonsandtrolls",
)