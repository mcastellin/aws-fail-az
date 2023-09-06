package awsapis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type ElbV2Api interface {
	ElbV2TagDescriptor
	ElbV2LoadBalancersDescriptor
	ElbV2SubnetSetter
	DescribeLoadBalancersPaginator
}

type ElbV2TagDescriptor interface {
	DescribeTags(context.Context,
		*elasticloadbalancingv2.DescribeTagsInput,
		...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error)
}

type ElbV2LoadBalancersDescriptor interface {
	DescribeLoadBalancers(context.Context,
		*elasticloadbalancingv2.DescribeLoadBalancersInput,
		...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
}

type ElbV2SubnetSetter interface {
	SetSubnets(context.Context,
		*elasticloadbalancingv2.SetSubnetsInput,
		...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.SetSubnetsOutput, error)
}

type DescribeLoadBalancersPaginator interface {
	NewDescribeLoadBalancersPaginator(
		params *elasticloadbalancingv2.DescribeLoadBalancersInput,
		optFn ...func(*elasticloadbalancingv2.Options)) DescribeLoadBalancersPager
}

type DescribeLoadBalancersPager interface {
	HasMorePages() bool
	NextPage(context.Context,
		...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
}

type AwsElbV2Api struct {
	client *elasticloadbalancingv2.Client
}

func (a *AwsElbV2Api) NewDescribeLoadBalancersPaginator(
	params *elasticloadbalancingv2.DescribeLoadBalancersInput,
	optFn ...func(*elasticloadbalancingv2.Options)) DescribeLoadBalancersPager {
	return elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(a.client, params)
}

func (a *AwsElbV2Api) DescribeTags(ctx context.Context,
	params *elasticloadbalancingv2.DescribeTagsInput,
	optFn ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
	return a.client.DescribeTags(ctx, params, optFn...)
}

func (a *AwsElbV2Api) DescribeLoadBalancers(ctx context.Context,
	params *elasticloadbalancingv2.DescribeLoadBalancersInput,
	optFn ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	return a.client.DescribeLoadBalancers(ctx, params, optFn...)
}

func (a *AwsElbV2Api) SetSubnets(ctx context.Context, params *elasticloadbalancingv2.SetSubnetsInput,
	optFn ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.SetSubnetsOutput, error) {
	return a.client.SetSubnets(ctx, params, optFn...)
}
