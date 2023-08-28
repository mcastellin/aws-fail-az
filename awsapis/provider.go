package awsapis

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// Creates a new provider from AWS configuration
func NewProviderFromConfig(cfg *aws.Config) AWSProvider {
	return AWSProviderImpl{
		awsConfig: cfg,
	}
}

type AWSProvider interface {
	NewDynamodbApi() DynamodbApi
	NewEc2Api() Ec2Api
	NewEcsApi() EcsApi
	NewAutoScalingApi() AutoScalingApi
}

type AWSProviderImpl struct {
	awsConfig *aws.Config
}

func (p AWSProviderImpl) NewDynamodbApi() DynamodbApi {
	return &AwsDynamodbApi{
		client: dynamodb.NewFromConfig(*p.awsConfig),
	}
}

func (p AWSProviderImpl) NewEc2Api() Ec2Api {
	return &AwsEc2Api{
		client: ec2.NewFromConfig(*p.awsConfig),
	}
}

func (p AWSProviderImpl) NewEcsApi() EcsApi {
	return &AwsEcsApi{
		client: ecs.NewFromConfig(*p.awsConfig),
	}
}

func (p AWSProviderImpl) NewAutoScalingApi() AutoScalingApi {
	return &AwsAutoScalingApi{
		client: autoscaling.NewFromConfig(*p.awsConfig),
	}
}
