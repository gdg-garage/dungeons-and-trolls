proto:
	bazel build //...
	cp ../bazel-bin/server/dungeonsandtrolls/proto/* dungeonsandtrolls/api/
	chmod u+w dungeonsandtrolls/*

update-deps:
	go get -t -u

clean:
	rm -rf ../bazel-*

test:
	go test ./...

download-generator:
	@cd generator; python3 download_latest.py
