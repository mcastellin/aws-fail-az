.PHONY: clean test build install mockgen

BUILD_VERSION=dev-snapshot

clean:
	go clean -testcache

test:
	go test ./...

build:
	go build -ldflags="-X main.BuildVersion=$(BUILD_VERSION)" -o bin/aws-fail-az cmd/*.go

install: build
	mkdir -p ~/bin/
	cp bin/aws-fail-az ~/bin/aws-fail-az
	chmod 700 ~/bin/aws-fail-az

mockgen:
	mockgen -source domain/asg.go -destination mock_domain/asg.go
	mockgen -source domain/ec2.go -destination mock_domain/ec2.go
