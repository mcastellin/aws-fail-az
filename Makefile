.PHONY: build

BUILD_VERSION=dev-snapshot

build:
	go build -ldflags="-X main.BuildVersion=$(BUILD_VERSION)" -o bin/aws-fail-az cmd/*.go

install: build
	mkdir -p ~/bin/
	cp bin/aws-fail-az ~/bin/aws-fail-az
	chmod 700 ~/bin/aws-fail-az
