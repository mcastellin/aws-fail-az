module github.com/mcastellin/aws-fail-az

go 1.21

require (
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.33
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.10.36
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.4.63
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.30.6
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.21.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.119.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.30.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.21.4
	github.com/mcastellin/aws-fail-az/awsapis v0.0.0-00010101000000-000000000000
	github.com/mcastellin/aws-fail-az/awsapis_mocks v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	go.uber.org/mock v0.3.0
	golang.org/x/exp v0.0.0-20230811145659-89c5cff77bcb
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.13.32 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.41 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.39 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.21.2 // indirect
	github.com/aws/smithy-go v1.14.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/mcastellin/aws-fail-az/awsapis => ./awsapis

replace github.com/mcastellin/aws-fail-az/awsapis_mocks => ./awsapis_mocks
