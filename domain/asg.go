package domain

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

func NewAutoScalingApi(provider *AWSProvider) AutoScalingApi {
	return &AwsAutoScalingApi{
		client: autoscaling.NewFromConfig(provider.GetConnection()),
	}
}

// Interfaces
type AutoScalingApi interface {
	AutoScalingDescriber
	AutoScalingGroupUpdater
	DescribeAutoScalingGroupsPaginator
}

type AutoScalingDescriber interface {
	DescribeAutoScalingGroups(ctx context.Context,
		params *autoscaling.DescribeAutoScalingGroupsInput,
		optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

type AutoScalingGroupUpdater interface {
	UpdateAutoScalingGroup(ctx context.Context,
		params *autoscaling.UpdateAutoScalingGroupInput,
		optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error)
}

type DescribeAutoScalingGroupsPaginator interface {
	NewDescribeAutoScalingGroupsPaginator(params *autoscaling.DescribeAutoScalingGroupsInput) DescribeAutoScalingGroupsPager
}

type DescribeAutoScalingGroupsPager interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

// Implementation
type AwsAutoScalingApi struct {
	client *autoscaling.Client
}

func (a *AwsAutoScalingApi) DescribeAutoScalingGroups(ctx context.Context,
	params *autoscaling.DescribeAutoScalingGroupsInput,
	optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {

	return a.client.DescribeAutoScalingGroups(ctx, params, optFns...)
}

func (a *AwsAutoScalingApi) UpdateAutoScalingGroup(ctx context.Context,
	params *autoscaling.UpdateAutoScalingGroupInput,
	optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	return a.client.UpdateAutoScalingGroup(ctx, params, optFns...)
}

func (a *AwsAutoScalingApi) NewDescribeAutoScalingGroupsPaginator(params *autoscaling.DescribeAutoScalingGroupsInput) DescribeAutoScalingGroupsPager {
	return autoscaling.NewDescribeAutoScalingGroupsPaginator(a.client, params)
}
