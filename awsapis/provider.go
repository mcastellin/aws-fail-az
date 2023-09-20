package awsapis

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

// Creates a new provider from AWS configuration
func NewProviderFromConfig(cfg *aws.Config) AWSProvider {
	return awsProviderImpl{
		awsConfig: cfg,
	}
}

type AWSProvider interface {
	NewDynamodbApi() DynamodbApi
	NewEc2Api() Ec2Api
	NewEcsApi() EcsApi
	NewAutoScalingApi() AutoScalingApi
	NewElbV2Api() ElbV2Api
}

type awsProviderImpl struct {
	awsConfig *aws.Config
}

func (p awsProviderImpl) NewDynamodbApi() DynamodbApi {
	return &AwsDynamodbApi{
		client: dynamodb.NewFromConfig(*p.awsConfig),
	}
}

func (p awsProviderImpl) NewEc2Api() Ec2Api {
	return &AwsEc2Api{
		client: ec2.NewFromConfig(*p.awsConfig),
	}
}

func (p awsProviderImpl) NewEcsApi() EcsApi {
	return &AwsEcsApi{
		client: ecs.NewFromConfig(*p.awsConfig),
	}
}

func (p awsProviderImpl) NewAutoScalingApi() AutoScalingApi {
	return &AwsAutoScalingApi{
		client: autoscaling.NewFromConfig(*p.awsConfig),
	}
}

func (p awsProviderImpl) NewElbV2Api() ElbV2Api {
	return &AwsElbV2Api{
		client: elasticloadbalancingv2.NewFromConfig(*p.awsConfig),
	}
}
