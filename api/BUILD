load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_openapiv2_compile")

gateway_openapiv2_compile(
    name = "openapi-spec",
    protos = ["//proto:dungeonsandtrolls_proto"],
    extra_protoc_files = ["//proto:service"],
    options = {
        "@rules_proto_grpc//grpc-gateway:openapiv2_plugin": [
            "grpc_api_configuration=proto/service.yml",
        ]
    },
)
