proto-plugin-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

proto:
	protoc --go_out=../server --go_opt=paths=source_relative --go-grpc_out=../server --go-grpc_opt=paths=source_relative --proto_path=../ ../spec/dungeonsandtrolls.proto