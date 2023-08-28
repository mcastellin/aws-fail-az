package asg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/awsutils"
)

func RestoreFromState(stateData []byte, provider *awsapis.AWSProvider) error {
	var state AutoScalingGroupState
	err := json.Unmarshal(stateData, &state)
	if err != nil {
		return err
	}

	return AutoScalingGroup{
		Provider:             *provider,
		AutoScalingGroupName: state.AutoScalingGroupName,
		stateSubnets:         state.Subnets,
	}.Restore()
}

func NewFromConfig(selector domain.ServiceSelector, provider *awsapis.AWSProvider) ([]domain.ConsistentStateService, error) {

	if selector.Type != RESOURCE_TYPE {
		return nil, fmt.Errorf("Unable to create AutoScalingGroup object from selector of type %s.", selector.Type)
	}

	var asgNames []string
	var err error

	err = selector.Validate()
	if err != nil {
		return nil, err
	}

	attributes, err := awsutils.TokenizeResourceFilter(selector.Filter, []string{"name"})
	if err != nil {
		return nil, err
	}

	if len(attributes) == 1 {
		asgNames = []string{attributes["name"]}

	} else if len(selector.Tags) > 0 {
		api := (*provider).NewAutoScalingApi()
		asgNames, err = filterAutoScalingGroupsByTags(api, selector.Tags)
		if err != nil {
			return nil, err
		}
	}

	objs := []domain.ConsistentStateService{}

	for _, name := range asgNames {
		objs = append(objs,
			AutoScalingGroup{
				Provider:             *provider,
				AutoScalingGroupName: name,
			})
	}

	return objs, nil
}

func filterAutoScalingGroupsByTags(api awsapis.AutoScalingApi, tags []domain.AWSTag) ([]string, error) {
	groupNames := []string{}

	paginator := api.NewDescribeAutoScalingGroupsPaginator(&autoscaling.DescribeAutoScalingGroupsInput{})
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
