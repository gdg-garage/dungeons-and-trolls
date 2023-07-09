load("@rules_proto//proto:defs.bzl", "proto_library")

package(default_visibility = ["//visibility:public"])

proto_library(
    name = "dungeonsandtrolls_proto",
    srcs = [
        "dungeonsandtrolls.proto",
    ],
    deps = [
        "@com_google_protobuf//:empty_proto",
        "@com_google_protobuf//:struct_proto",
    ],
)

proto_library(
    name = "dungeonsandtrolls_annotated_proto",
    srcs = [
        "dungeonsandtrolls_annotated.proto",
    ],
    deps = [
        "@go_googleapis//google/api:annotations_proto",
    ],
)

filegroup(
  name = 'service',
  srcs = ["service.yml"],
)
