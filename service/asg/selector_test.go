package asg

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/awsapis_mocks"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFilterAsgByTagsShouldNotMatchResults(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()
	mockApi := awsapis_mocks.NewMockAutoScalingApi(ctrl)

	pages := [][]types.AutoScalingGroup{{
		{
			AutoScalingGroupName: aws.String("asg-name-test"),
			Tags: []types.TagDescription{
				{Key: aws.String("Application"), Value: aws.String("myapp")},
				{Key: aws.String("Application"), Value: aws.String("test")},
			},
		},
	}}
	mockPager := createDescribeAsgPaginator(ctrl, pages)

	mockApi.EXPECT().
		NewDescribeAutoScalingGroupsPaginator(gomock.Any()).
		Times(1).
		Return(mockPager)

	filter := []domain.AWSTag{{
		Name:  "Application",
		Value: "myapp",
	}, {
		Name:  "Environment",
		Value: "live",
	}}

	result, err := filterAutoScalingGroupsByTags(mockApi, filter)

	assert.Len(t, result, 0)
	assert.Nil(t, err)

}

func TestFilterAsgByTagsShouldMatchResultsInAllPages(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()
	mockApi := awsapis_mocks.NewMockAutoScalingApi(ctrl)

	pages := [][]types.AutoScalingGroup{
		{{
			AutoScalingGroupName: aws.String("asg-name-test"),
			Tags: []types.TagDescription{
				{Key: aws.String("Application"), Value: aws.String("myapp")},
				{Key: aws.String("Application"), Value: aws.String("test")},
			},
		}},
		{{
			AutoScalingGroupName: aws.String("asg-name-live"),
			Tags: []types.TagDescription{
				{Key: aws.String("Application"), Value: aws.String("myapp")},
				{Key: aws.String("Application"), Value: aws.String("live")},
			},
		}},
	}
	mockPager := createDescribeAsgPaginator(ctrl, pages)

	mockApi.EXPECT().
		NewDescribeAutoScalingGroupsPaginator(gomock.Any()).
		Times(1).
		DoAndReturn(func(_ *autoscaling.DescribeAutoScalingGroupsInput) awsapis.DescribeAutoScalingGroupsPager {
			return mockPager
		})

	filter := []domain.AWSTag{{
		Name:  "Application",
		Value: "myapp",
	}}

	result, err := filterAutoScalingGroupsByTags(mockApi, filter)
	t.Log(result)

	assert.Equal(t, []string{"asg-name-test", "asg-name-live"}, result)
	assert.Nil(t, err)
}

func createDescribeAsgPaginator(ctrl *gomock.Controller, pages [][]types.AutoScalingGroup) *awsapis_mocks.MockDescribeAutoScalingGroupsPager {
	mockPager := awsapis_mocks.NewMockDescribeAutoScalingGroupsPager(ctrl)

	gomock.InOrder(
		mockPager.EXPECT().HasMorePages().Times(len(pages)).Return(true),
		mockPager.EXPECT().HasMorePages().Times(1).Return(false),
	)
	calls := []*gomock.Call{}
	for idx := range pages {
		c := mockPager.EXPECT().NextPage(gomock.Any()).Times(1).Return(
			&autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: pages[idx],
			}, nil)
		calls = append(calls, c)
	}
	gomock.InOrder(calls...)

	return mockPager
}
