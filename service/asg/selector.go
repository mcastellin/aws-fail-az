package asg

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/mcastellin/aws-fail-az/domain"
)

func NewFromConfig(selector domain.ServiceSelector, provider *domain.AWSProvider) ([]domain.ConsistentStateService, error) {

	if selector.Type != RESOURCE_TYPE {
		return nil, fmt.Errorf("Unable to create AutoScalingGroup object from selector of type %s.", selector.Type)
	}

	var asgNames []string
	var err error

	err = domain.ValidateServiceSelector(selector)
	if err != nil {
		return nil, err
	}

	if selector.Filter != "" {
		tokens := strings.Split(selector.Filter, "=")
		key := tokens[0]
		value := tokens[1]

		if key == "name" {
			asgNames = []string{value}
		} else {
			return nil, fmt.Errorf("Unrecognized key %s for type %s", key, RESOURCE_TYPE)
		}
	} else if len(selector.Tags) > 0 {
		client := autoscaling.NewFromConfig(provider.GetConnection())
		asgNames, err = FilterAutoScalingGroupsByTags(client, selector.Tags)
		if err != nil {
			return nil, err
		}
	}

	objs := []domain.ConsistentStateService{}

	for _, name := range asgNames {
		objs = append(objs,
			AutoScalingGroup{
				Provider:             provider,
				AutoScalingGroupName: name,
			})
	}

	return objs, nil
}

func FilterAutoScalingGroupsByTags(client *autoscaling.Client, tags []domain.AWSTag) ([]string, error) {
	groupNames := []string{}

	paginator := autoscaling.NewDescribeAutoScalingGroupsPaginator(client, &autoscaling.DescribeAutoScalingGroupsInput{})
	for paginator.HasMorePages() {
		response, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, group := range response.AutoScalingGroups {
			allMatch := len(group.Tags) >= len(tags)
			for _, filterTag := range tags {
				match := false
				for _, resourceTag := range group.Tags {
					if filterTag.Name == *resourceTag.Key && filterTag.Value == *resourceTag.Value {
						match = true
					}
				}
				allMatch = allMatch && match
			}
			if allMatch {
				groupNames = append(groupNames, *group.AutoScalingGroupName)
			}
		}
	}

	return groupNames, nil
}
