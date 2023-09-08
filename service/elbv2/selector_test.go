package elbv2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/mock_awsapis"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFilterLoadBalancersByTagsShouldMatchInAllPages(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	arns := []string{
		"arn:aws:elasticloadbalancing:us-east-1:000000000000:loadbalancer/app/test-alb-1/xxxxxxxxxxxxxxx",
		"arn:aws:elasticloadbalancing:us-east-1:000000000000:loadbalancer/app/test-alb-2/xxxxxxxxxxxxxxx",
		"arn:aws:elasticloadbalancing:us-east-1:000000000000:loadbalancer/app/test-alb-3/xxxxxxxxxxxxxxx",
		"arn:aws:elasticloadbalancing:us-east-1:000000000000:loadbalancer/app/test-alb-4/xxxxxxxxxxxxxxx",
	}

	pages := [][]types.LoadBalancer{
		{{
			LoadBalancerArn: aws.String(arns[0]),
		}},
		{
			{LoadBalancerArn: aws.String(arns[1])},
			{LoadBalancerArn: aws.String(arns[2])},
			{LoadBalancerArn: aws.String(arns[3])},
		},
	}
	describeLoadBalancersPager := createDescribeLoadBalancersPager(ctrl, pages)

	mockApi := mock_awsapis.NewMockElbV2Api(ctrl)
	mockApi.EXPECT().NewDescribeLoadBalancersPaginator(gomock.Any()).Times(1).Return(describeLoadBalancersPager)

	gomock.InOrder(
		mockApi.EXPECT().DescribeTags(gomock.Any(), gomock.Any()).Times(1).
			Return(&elasticloadbalancingv2.DescribeTagsOutput{TagDescriptions: []types.TagDescription{
				{
					ResourceArn: aws.String(arns[0]),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("live")},
						{Key: aws.String("Application"), Value: aws.String("test")},
						{Key: aws.String("Other"), Value: aws.String("tag")},
						{Key: aws.String("Name"), Value: aws.String("somename")},
					},
				},
			}}, nil),
		mockApi.EXPECT().DescribeTags(gomock.Any(), gomock.Any()).Times(1).
			Return(&elasticloadbalancingv2.DescribeTagsOutput{TagDescriptions: []types.TagDescription{
				{
					ResourceArn: aws.String(arns[1]),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("live")},
						{Key: aws.String("Application"), Value: aws.String("test")},
						{Key: aws.String("Other"), Value: aws.String("tag")},
						{Key: aws.String("Name"), Value: aws.String("somename")},
					},
				},
				{
					ResourceArn: aws.String(arns[2]),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("live")},
						{Key: aws.String("Application"), Value: aws.String("nomatch")},
						{Key: aws.String("Name"), Value: aws.String("somename")},
					},
				},
				{
					ResourceArn: aws.String(arns[3]),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("live")},
						{Key: aws.String("Application"), Value: aws.String("test")},
					},
				},
			}}, nil),
	)

	mockProvider := mock_awsapis.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewElbV2Api().AnyTimes().Return(mockApi)

	config := domain.TargetSelector{
		Type: RESOURCE_TYPE,
		Tags: []domain.AWSTag{{Name: "Environment", Value: "live"}, {Name: "Application", Value: "test"}},
	}
	results, err := NewFromConfig(config, mockProvider)

	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, arns[0], results[0].(*LoadBalancer).Name)
	assert.Equal(t, arns[1], results[1].(*LoadBalancer).Name)
	assert.Equal(t, arns[3], results[2].(*LoadBalancer).Name)
}

func createDescribeLoadBalancersPager(ctrl *gomock.Controller, pages [][]types.LoadBalancer) *mock_awsapis.MockDescribeLoadBalancersPager {
	pager := mock_awsapis.NewMockDescribeLoadBalancersPager(ctrl)

	gomock.InOrder(
		pager.EXPECT().HasMorePages().Times(len(pages)).Return(true),
		pager.EXPECT().HasMorePages().Times(1).Return(false),
	)

	calls := make([]*gomock.Call, len(pages))
	for idx := range pages {
		calls[idx] = pager.EXPECT().NextPage(gomock.Any()).Times(1).
			Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
				LoadBalancers: pages[idx],
			}, nil)
	}
	gomock.InOrder(calls...)

	return pager
}
