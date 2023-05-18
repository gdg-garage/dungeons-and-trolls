proto:
	bazel build //...
	cp ../bazel-bin/server/dungeonsandtrolls/* dungeonsandtrolls/
	chmod u+w dungeonsandtrolls/*

update-deps:
	go get -t -u

clean:
	rm -rf ../bazel-*