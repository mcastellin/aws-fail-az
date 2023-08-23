.PHONY: clean test build install mockgen

BUILD_VERSION=dev-snapshot

clean:
	go clean -testcache

test:
	go test ./...

build: clean
	go build -ldflags="-X main.BuildVersion=$(BUILD_VERSION)" -o bin/aws-fail-az cmd/*.go

install: build
	mkdir -p ~/bin/
	cp bin/aws-fail-az ~/bin/aws-fail-az
	chmod 700 ~/bin/aws-fail-az

# Auto-generate AWS api mocks for unit testing
# IMPORTANT!! Run this target every time you need to modify the `domain` package
mockgen:
	mockgen -source awsapis/asg.go -destination mock_awsapis/asg.go
	mockgen -source awsapis/ec2.go -destination mock_awsapis/ec2.go
	mockgen -source awsapis/ecs.go -destination mock_awsapis/ecs.go
	mockgen -source awsapis/dynamodb.go -destination mock_awsapis/dynamodb.go
