load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_grpc_compile")

gateway_grpc_compile(
    name = "dungeonsandtrolls_gw",
    protos = ["//proto:dungeonsandtrolls_proto"],
    output_mode = "NO_PREFIX",
    prefix_path = "dungeonsandtrolls",
)