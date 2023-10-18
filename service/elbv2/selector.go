package elbv2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/awsutils"
)

func DescribeElbv2LoadBalancersState(stateData []byte) (string, error) {
	var state LoadBalancerState
	err := json.Unmarshal(stateData, &state)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("- LoadBalancerName: %s", state.LoadBalancerName), nil
}

func RestoreElbv2LoadBalancersFromState(stateData []byte, provider awsapis.AWSProvider) error {
	var state LoadBalancerState
	err := json.Unmarshal(stateData, &state)
	if err != nil {
		return err
	}

	resource := LoadBalancer{
		Provider:     provider,
		Name:         state.LoadBalancerName,
		stateSubnets: state.Subnets,
	}
	return resource.Restore()
}

func NewElbv2LoadBalancerFaultFromConfig(selector domain.TargetSelector, provider awsapis.AWSProvider) ([]domain.ConsistentStateResource, error) {

	if selector.Type != domain.ResourceTypeElbv2LoadBalancer {
		return nil, fmt.Errorf("Unable to create LoadBalancer object from selector of type %s.", selector.Type)
	}

	var lbNames []string
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
		lbNames = []string{attributes["name"]}
	} else if len(selector.Tags) > 0 {
		api := provider.NewElbV2Api()

		lbNames, err = filterLoadBalancersByTag(api, selector.Tags)
		if err != nil {
			return nil, err
		}
	}

	objs := make([]domain.ConsistentStateResource, len(lbNames))
	for idx := range lbNames {
		objs[idx] = &LoadBalancer{
			Provider: provider,
			Name:     lbNames[idx],
		}
	}

	return objs, nil
}

func filterLoadBalancersByTag(api awsapis.ElbV2Api, tags []domain.AWSTag) ([]string, error) {
	lbNames := []string{}

	paginator := api.NewDescribeLoadBalancersPaginator(
		&elasticloadbalancingv2.DescribeLoadBalancersInput{})

	for paginator.HasMorePages() {
		response, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		if len(response.LoadBalancers) == 0 {
			continue
		}

		resourceArns := make([]string, len(response.LoadBalancers))
		for idx, lb := range response.LoadBalancers {
			resourceArns[idx] = *lb.LoadBalancerArn
		}

		describeTagsOutput, err := api.DescribeTags(context.TODO(),
			&elasticloadbalancingv2.DescribeTagsInput{ResourceArns: resourceArns})
		if err != nil {
			return nil, err
		}

		for _, descriptor := range describeTagsOutput.TagDescriptions {
			if resourceTagsMatchFilters(descriptor, tags) {
				lbNames = append(lbNames, *descriptor.ResourceArn)
			}
		}
	}

	return lbNames, nil
}

func resourceTagsMatchFilters(tagDescriptor types.TagDescription, filterTags []domain.AWSTag) bool {
	allMatch := len(tagDescriptor.Tags) >= len(filterTags)
	for _, filterTag := range filterTags {
		match := false
		for _, resourceTag := range tagDescriptor.Tags {
			if *resourceTag.Key == filterTag.Name && *resourceTag.Value == filterTag.Value {
				match = true
			}
		}
		allMatch = allMatch && match
	}
	return allMatch
}
