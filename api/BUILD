load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_openapiv2_compile")

gateway_openapiv2_compile(
    name = "openapi-spec",
    protos = ["//proto:dungeonsandtrolls_proto"],
    extra_protoc_files = [
        "//proto:service",
        "//api:openapi_options",
    ],
    options = {
        "@rules_proto_grpc//grpc-gateway:openapiv2_plugin": [
            "grpc_api_configuration=proto/service.yml",
            "openapi_configuration=api/openapi_options.yaml",
            "proto3_optional_nullable=true"
        ]
    },
)

filegroup(
  name = 'openapi_options',
  srcs = ["openapi_options.yaml"],
)