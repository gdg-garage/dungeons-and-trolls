proto-plugin-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

proto:
	# protoc --go_out=../server/dungeonsandtrolls --go_opt=paths=source_relative \
    #        --go-grpc_out=../server/dungeonsandtrolls --go-grpc_opt=paths=source_relative \
    #        --grpc-gateway_out=../server/dungeonsandtrolls --grpc-gateway_opt paths=source_relative \
    #        --proto_path=../spec ../proto/dungeonsandtrolls.proto
	bazel build //...
	cp ../bazel-bin/server/dungeonsandtrolls/* dungeonsandtrolls/
	chmod u+w dungeonsandtrolls/*

update-deps:
	go get -t -u

clean:
	rm -rf ../bazel-*