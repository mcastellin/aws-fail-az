.PHONY: clean test build install mockgen

BUILD_VERSION=dev-snapshot

clean:
	go clean -testcache

test:
	go test ./... -v

build: clean
	go build -ldflags="-X main.BuildVersion=$(BUILD_VERSION)" -o bin/aws-fail-az cmd/*.go

install: build
	mkdir -p ~/bin/
	cp bin/aws-fail-az ~/bin/aws-fail-az
	chmod 700 ~/bin/aws-fail-az

# Auto-generate AWS api mocks for unit testing
# IMPORTANT!! Run this target every time you need to modify the `domain` package
mockgen: ./awsapis/*.go
	@for file in $^; do \
			echo Generating mocks for $$file; \
			mockgen -source $$file -destination mock_awsapis/$$(basename $$file); \
		done
