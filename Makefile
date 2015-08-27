NAME=port-checker
OS=linux darwin

.PHONY: deps test build install

deps:
	go get github.com/tools/godep
	go get github.com/mitchellh/gox/...

build:
	mkdkr -p build/
	gox -os="$OS" -arch="amd64" -output="build/{{.OS}}/$(NAME)"

install: build
	install build/$(shell uname -s)/$(NAME) /usr/local/bin
