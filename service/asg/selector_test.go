package asg

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/mock_domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFilterAsgByTagsShouldNotMatchResults(t *testing.T) {

	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()
	mockApi := mock_domain.NewMockAutoScalingApi(ctrl)
	mockPager := mock_domain.NewMockDescribeAutoScalingGroupsPager(ctrl)

	mockApi.EXPECT().
		NewDescribeAutoScalingGroupsPaginator(gomock.Any()).
		Times(1).
		DoAndReturn(func(_ *autoscaling.DescribeAutoScalingGroupsInput) domain.DescribeAutoScalingGroupsPager {
			return mockPager
		})

	gomock.InOrder(
		mockPager.EXPECT().
			HasMorePages().
			Times(1).
			DoAndReturn(func() bool {
				return true
			}),
		mockPager.EXPECT().
			HasMorePages().
			Times(1).
			DoAndReturn(func() bool {
				return false
			}),
	)

	mockPager.EXPECT().
		NextPage(gomock.Any()).
		Times(1).
		DoAndReturn(func(_ context.Context, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
			out := &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []types.AutoScalingGroup{{
					AutoScalingGroupName: aws.String("asg-name-test"),
					Tags: []types.TagDescription{
						{Key: aws.String("Application"), Value: aws.String("myapp")},
						{Key: aws.String("Application"), Value: aws.String("test")},
					},
				}},
			}

			return out, nil
		})

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
	mockApi := mock_domain.NewMockAutoScalingApi(ctrl)
	mockPager := mock_domain.NewMockDescribeAutoScalingGroupsPager(ctrl)

	mockApi.EXPECT().
		NewDescribeAutoScalingGroupsPaginator(gomock.Any()).
		Times(1).
		DoAndReturn(func(_ *autoscaling.DescribeAutoScalingGroupsInput) domain.DescribeAutoScalingGroupsPager {
			return mockPager
		})

	gomock.InOrder(
		mockPager.EXPECT().
			HasMorePages().
			Times(2).
			DoAndReturn(func() bool {
				return true
			}),
		mockPager.EXPECT().
			HasMorePages().
			Times(1).
			DoAndReturn(func() bool {
				return false
			}),
	)

	gomock.InOrder(
		mockPager.EXPECT().
			NextPage(gomock.Any()).
			Times(1).
			DoAndReturn(func(_ context.Context, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
				out := &autoscaling.DescribeAutoScalingGroupsOutput{
					AutoScalingGroups: []types.AutoScalingGroup{{
						AutoScalingGroupName: aws.String("asg-name-test"),
						Tags: []types.TagDescription{
							{Key: aws.String("Application"), Value: aws.String("myapp")},
							{Key: aws.String("Application"), Value: aws.String("test")},
						},
					}},
				}

				return out, nil
			}),
		mockPager.EXPECT().
			NextPage(gomock.Any()).
			Times(1).
			DoAndReturn(func(_ context.Context, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
				out := &autoscaling.DescribeAutoScalingGroupsOutput{
					AutoScalingGroups: []types.AutoScalingGroup{{
						AutoScalingGroupName: aws.String("asg-name-live"),
						Tags: []types.TagDescription{
							{Key: aws.String("Application"), Value: aws.String("myapp")},
							{Key: aws.String("Application"), Value: aws.String("live")},
						},
					}},
				}

				return out, nil
			}),
	)

	filter := []domain.AWSTag{{
		Name:  "Application",
		Value: "myapp",
	}}

	result, err := filterAutoScalingGroupsByTags(mockApi, filter)
	t.Log(result)

	assert.Equal(t, []string{"asg-name-test", "asg-name-live"}, result)
	assert.Nil(t, err)
}
