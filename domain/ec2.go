package domain

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func NewEc2Api(provider *AWSProvider) Ec2Api {
	return &AwsEc2Api{
		client: ec2.NewFromConfig(provider.GetConnection()),
	}
}

// Interfaces
type Ec2Api interface {
	Ec2SubnetsDescriptor
	Ec2InstanceTerminator
}

type Ec2SubnetsDescriptor interface {
	DescribeSubnets(ctx context.Context,
		params *ec2.DescribeSubnetsInput,
		optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
}

type Ec2InstanceTerminator interface {
	TerminateInstances(ctx context.Context,
		params *ec2.TerminateInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
}

// Implementation
type AwsEc2Api struct {
	client *ec2.Client
}

func (a *AwsEc2Api) DescribeSubnets(ctx context.Context,
	params *ec2.DescribeSubnetsInput,
	optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {

	return a.client.DescribeSubnets(ctx, params, optFns...)
}

func (a *AwsEc2Api) TerminateInstances(ctx context.Context,
	params *ec2.TerminateInstancesInput,
	optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {

	return a.client.TerminateInstances(ctx, params, optFns...)
}
