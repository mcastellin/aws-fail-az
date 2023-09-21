package elbv2

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/mcastellin/aws-fail-az/awsapis_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCheckFailNotEnoughSubnets(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := awsapis_mocks.NewMockElbV2Api(ctrl)
	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
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

	result, err := (&LoadBalancer{
		Provider: mockProvider,
		Name:     "test-alb",
	}).Check()

	assert.NotNil(t, err)
	assert.False(t, result)
}

func TestCheckPass(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := awsapis_mocks.NewMockElbV2Api(ctrl)
	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
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

	result, err := (&LoadBalancer{
		Provider: mockProvider,
		Name:     "test-alb",
	}).Check()

	assert.Nil(t, err)
	assert.True(t, result)
}

func TestDescribeLoadBalancerWithArn(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	const arn string = "arn:aws:elasticloadbalancing:us-east-1:000000000000:loadbalancer/app/test-alb-1/xxxxxxxxxxxxxxx"

	mockApi := awsapis_mocks.NewMockElbV2Api(ctrl)
	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewElbV2Api().AnyTimes().Return(mockApi)

	matcher := describeLoadBalancersMatcher{&elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{arn},
	}}
	mockApi.EXPECT().DescribeLoadBalancers(gomock.Any(), matcher).Times(1).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{{
				AvailabilityZones: []types.AvailabilityZone{
					{SubnetId: aws.String("s-1111")},
					{SubnetId: aws.String("s-2222")},
					{SubnetId: aws.String("s-3333")},
				},
			}},
		}, nil)

	result, err := (&LoadBalancer{
		Provider: mockProvider,
		Name:     arn,
	}).Check()

	assert.Nil(t, err)
	assert.True(t, result)
}

func TestDescribeLoadBalancerWithName(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := awsapis_mocks.NewMockElbV2Api(ctrl)
	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewElbV2Api().AnyTimes().Return(mockApi)

	matcher := describeLoadBalancersMatcher{&elasticloadbalancingv2.DescribeLoadBalancersInput{
		Names: []string{"alb-name"},
	}}
	mockApi.EXPECT().DescribeLoadBalancers(gomock.Any(), matcher).Times(1).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{{
				AvailabilityZones: []types.AvailabilityZone{
					{SubnetId: aws.String("s-1111")},
					{SubnetId: aws.String("s-2222")},
					{SubnetId: aws.String("s-3333")},
				},
			}},
		}, nil)

	result, err := (&LoadBalancer{
		Provider: mockProvider,
		Name:     "alb-name",
	}).Check()

	assert.Nil(t, err)
	assert.True(t, result)
}

// Matchers
type describeLoadBalancersMatcher struct {
	x *elasticloadbalancingv2.DescribeLoadBalancersInput
}

func (m describeLoadBalancersMatcher) Matches(x interface{}) bool {
	if y, ok := x.(*elasticloadbalancingv2.DescribeLoadBalancersInput); ok {
		return reflect.DeepEqual(m.x, y)
	}
	return false
}

func (m describeLoadBalancersMatcher) String() string {
	return fmt.Sprintf("%v", m.x)
}
