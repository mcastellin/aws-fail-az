module github.com/mcastellin/aws-fail-az/awsapis_mocks

go 1.21

require (
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.30.6
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.21.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.119.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.30.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.21.4
	github.com/mcastellin/aws-fail-az/awsapis v0.0.0-00010101000000-000000000000
	go.uber.org/mock v0.3.0
)

require (
	github.com/aws/aws-sdk-go-v2 v1.21.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.41 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.35 // indirect
	github.com/aws/smithy-go v1.14.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)

replace github.com/mcastellin/aws-fail-az => ../

replace github.com/mcastellin/aws-fail-az/awsapis => ../awsapis
