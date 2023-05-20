load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_openapiv2_compile")

# gazelle:proto_plugin protoc-gen-grpc-gateway option logtostderr=true
# gazelle:proto_plugin protoc-gen-grpc-gateway option generate_unbound_methods=true

# gazelle:proto_plugin protoc-gen-grpc-openapiv2 implementation grpc-ecosystem:grpc-gateway:protoc-gen-grpc-openapiv2
# gazelle:proto_plugin protoc-gen-grpc-openapiv2 option logtostderr=true
# gazelle:proto_plugin protoc-gen-grpc-openapiv2 option use_go_templates=true
# gazelle:proto_plugin protoc-gen-grpc-openapiv2 option grpc_api_configuration=proto/service.yml

gateway_openapiv2_compile(
    name = "openapi-spec",
    protos = ["//proto:dungeonsandtrolls_proto"],
    extra_protoc_files = ["//proto:service"],
    # prefix_path = "dungeonsandtrolls",
    # options = {
    #     "@rules_proto_grpc//grpc-gateway:openapiv2_plugin": [
    #         "grpc_api_configuration=proto/service.yml",
    #     ]
    # },
    extra_protoc_args = ["--openapiv2_plugin_opt=grpc_api_configuration=proto/service.yml"]
)

# load("@com_github_grpc_ecosystem_grpc_gateway//protoc-gen-openapiv2:defs.bzl", "protoc_gen_openapiv2")
load("@com_github_grpc_ecosystem_grpc_gateway_v2//protoc-gen-openapiv2:defs.bzl", "protoc_gen_openapiv2")


protoc_gen_openapiv2(
    name = "openapi",
    proto = "//proto:dungeonsandtrolls_proto",
    # proto = "//proto:dungeonsandtrolls_annotated_proto",
    single_output = True,
    generate_unbound_methods = True,
    # grpc_api_configuration = "proto/service.yml",
    grpc_api_configuration = "//proto:service",

    # output_format = "yaml"
    # deps = ["//proto:service"],
)

# proto_compile(
#     name = "pluriproto",
#     protos = [":protos"],
#     langs = [
#         "//go",
#         "//python",
#     ],
#     verbose = 1,
#     with_grpc = True,
# )


# load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")



# go_proto_library(
#     name = "foo_go_proto",
#     importpath = "github.com/example/protos/foo_proto",
#     protos = ["//proto:dungeonsandtrolls_proto",],
#     visibility = ["//visibility:public"],
# )


# load("@org_pubref_rules_protobuf//grpc_gateway:rules.bzl", "grpc_gateway_swagger_compile")


# load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_binary", "go_test")
# load(
#     "//grpc_gateway:rules.bzl",
#     "grpc_gateway_proto_compile",
#     "grpc_gateway_proto_library",
#     "grpc_gateway_swagger_compile",
#     "grpc_gateway_binary",
#     "GRPC_GATEWAY_DEPS",
# )

# grpc_gateway_proto_library(
#     name = "gateway",
#     protos = ["//proto:dungeonsandtrolls_proto",],
# )

# grpc_gateway_swagger_compile(
#     name = "swagger",
#     protos = ["//proto:dungeonsandtrolls_proto",],
# )

# load("@build_stack_rules_proto//rules:proto_compile.bzl", "proto_compile")

# proto_compile(
#     name = "thing_go_compile",
#     output_mappings = ["thing.pb.go=github.com/stackb/rules_proto/example/thing/thing.pb.go"],
#     outputs = ["thing.pb.go"],
#     plugins = ["@build_stack_rules_proto//plugin/golang/protobuf:protoc-gen-go"],
#     proto = "//proto:dungeonsandtrolls_proto",
# )