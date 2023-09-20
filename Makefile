.PHONY: tidy clean test lint build install mockgen

BUILD_VERSION=dev-snapshot

tidy:
	@go mod tidy
	@cd awsapis/ && go mod tidy

clean:
	@go clean -testcache

test:
	@go test ./... -v

lint:
	@golangci-lint run

build: clean
	go build -ldflags="-X main.BuildVersion=$(BUILD_VERSION)" -o bin/aws-fail-az main.go

install: build
	@mkdir -p ~/bin/
	@cp bin/aws-fail-az ~/bin/aws-fail-az
	@chmod 700 ~/bin/aws-fail-az

# Auto-generate AWS api mocks for unit testing
# IMPORTANT!! Run this target every time you need to modify the `awsapis` module
mockgen: ./awsapis/*.go
	@for file in $^; do \
			echo Generating mocks for $$file; \
			mockgen -source $$file \
				-package awsapis_mocks \
				-destination awsapis_mocks/$$(basename $$file); \
		done
