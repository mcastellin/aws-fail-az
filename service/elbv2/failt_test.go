package elbv2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/mcastellin/aws-fail-az/mock_awsapis"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCheckFailNotEnoughSubnets(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockElbV2Api(ctrl)
	mockProvider := mock_awsapis.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewElbV2Api().AnyTimes().Return(mockApi)

	mockApi.EXPECT().DescribeLoadBalancers(gomock.Any(), gomock.Any()).Times(1).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{{
				AvailabilityZones: []types.AvailabilityZone{
					{SubnetId: aws.String("s-1111")},
					{SubnetId: aws.String("s-2222")},
				},
			}},
		}, nil)

	result, err := LoadBalancer{
		Provider: mockProvider,
		Name:     "test-alb",
	}.Check()

	assert.NotNil(t, err)
	assert.False(t, result)
}

func TestCheckPass(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockElbV2Api(ctrl)
	mockProvider := mock_awsapis.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewElbV2Api().AnyTimes().Return(mockApi)

	mockApi.EXPECT().DescribeLoadBalancers(gomock.Any(), gomock.Any()).Times(1).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{{
				AvailabilityZones: []types.AvailabilityZone{
					{SubnetId: aws.String("s-1111")},
					{SubnetId: aws.String("s-2222")},
					{SubnetId: aws.String("s-3333")},
				},
			}},
		}, nil)

	result, err := LoadBalancer{
		Provider: mockProvider,
		Name:     "test-alb",
	}.Check()

	assert.Nil(t, err)
	assert.True(t, result)
}
