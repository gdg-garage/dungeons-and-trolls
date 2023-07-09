load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_grpc_compile")

gateway_grpc_compile(
    name = "dungeonsandtrolls_gw",
    protos = ["//proto:dungeonsandtrolls_proto"],
    extra_protoc_files = ["//proto:service"],
    output_mode = "NO_PREFIX",
    prefix_path = "dungeonsandtrolls",
    options = {
        "@rules_proto_grpc//grpc-gateway:grpc_gateway_plugin": [
            "grpc_api_configuration=proto/service.yml",
        ]
    },
)