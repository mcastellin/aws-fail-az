package asg

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/mcastellin/aws-fail-az/domain"
)

func NewApiFromConfig(provider *domain.AWSProvider) apiConfigImpl {
	return apiConfigImpl{
		ec2Client:         ec2.NewFromConfig(provider.GetConnection()),
		autoscalingClient: autoscaling.NewFromConfig(provider.GetConnection()),
	}
}

type APIConfig interface {
	NewDescribeAutoScalingGroupsPager(*autoscaling.DescribeAutoScalingGroupsInput) DescribeAutoScalingGroupsPager
}

type DescribeAutoScalingGroupsPager interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

type apiConfigImpl struct {
	ec2Client         *ec2.Client
	autoscalingClient *autoscaling.Client
}

func (c apiConfigImpl) NewDescribeAutoScalingGroupsPager(
	input *autoscaling.DescribeAutoScalingGroupsInput) DescribeAutoScalingGroupsPager {
	return autoscaling.NewDescribeAutoScalingGroupsPaginator(c.autoscalingClient, input)
}

// Mock implementation for the APIConfig interface
type mockAPIConfig struct {
	DescribeAutoScalingGroupsPager DescribeAutoScalingGroupsPager
}

func (c mockAPIConfig) NewDescribeAutoScalingGroupsPager(
	input *autoscaling.DescribeAutoScalingGroupsInput) DescribeAutoScalingGroupsPager {
	return c.DescribeAutoScalingGroupsPager
}

// Mock implementation for the DescribeAutoScalingGroupPager interface
type mockDescribeAutoScalingGroupPager struct {
	PageNum int
	Pages   []*autoscaling.DescribeAutoScalingGroupsOutput
}

func (m *mockDescribeAutoScalingGroupPager) HasMorePages() bool {
	return m.PageNum < len(m.Pages)
}

func (m *mockDescribeAutoScalingGroupPager) NextPage(ctx context.Context, f ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	if m.PageNum >= len(m.Pages) {
		return nil, fmt.Errorf("No more pages")
	}
	out := m.Pages[m.PageNum]
	m.PageNum++
	return out, nil
}
