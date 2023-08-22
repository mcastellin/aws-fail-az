package asg

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/stretchr/testify/assert"
)

func TestFilterAsgByTagsShouldNotMatchResults(t *testing.T) {

	pager := &mockDescribeAutoScalingGroupPager{
		Pages: []*autoscaling.DescribeAutoScalingGroupsOutput{{
			AutoScalingGroups: []types.AutoScalingGroup{{
				AutoScalingGroupName: aws.String("asg-name-test"),
				Tags: []types.TagDescription{
					{Key: aws.String("Application"), Value: aws.String("myapp")},
					{Key: aws.String("Application"), Value: aws.String("test")},
				},
			}},
		}},
	}
	apiConfig := mockAPIConfig{DescribeAutoScalingGroupsPager: pager}

	filter := []domain.AWSTag{{
		Name:  "Application",
		Value: "myapp",
	}, {
		Name:  "Environment",
		Value: "live",
	}}

	result, err := filterAutoScalingGroupsByTags(apiConfig, filter)

	assert.Len(t, result, 0)
	assert.Nil(t, err)

}

func TestFilterAsgByTagsShouldMatchResultsInAllPages(t *testing.T) {

	pager := &mockDescribeAutoScalingGroupPager{
		Pages: []*autoscaling.DescribeAutoScalingGroupsOutput{{
			AutoScalingGroups: []types.AutoScalingGroup{{
				AutoScalingGroupName: aws.String("asg-name-test"),
				Tags: []types.TagDescription{
					{Key: aws.String("Application"), Value: aws.String("myapp")},
					{Key: aws.String("Application"), Value: aws.String("test")},
				},
			}},
		}, {
			AutoScalingGroups: []types.AutoScalingGroup{{
				AutoScalingGroupName: aws.String("asg-name-live"),
				Tags: []types.TagDescription{
					{Key: aws.String("Application"), Value: aws.String("myapp")},
					{Key: aws.String("Application"), Value: aws.String("live")},
				},
			}},
		}},
	}
	apiConfig := mockAPIConfig{DescribeAutoScalingGroupsPager: pager}

	filter := []domain.AWSTag{{
		Name:  "Application",
		Value: "myapp",
	}}

	result, err := filterAutoScalingGroupsByTags(apiConfig, filter)
	t.Log(result)

	assert.Equal(t, []string{"asg-name-test", "asg-name-live"}, result)
	assert.Nil(t, err)
}
