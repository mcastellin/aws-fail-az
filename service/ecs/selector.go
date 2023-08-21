package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/mcastellin/aws-fail-az/domain"
)

func RestoreFromState(stateData []byte, provider *domain.AWSProvider) error {
	var state ECSServiceState
	err := json.Unmarshal(stateData, &state)
	if err != nil {
		return err
	}

	return ECSService{
		Provider:     provider,
		ClusterArn:   state.ClusterArn,
		ServiceName:  state.ServiceName,
		stateSubnets: state.Subnets,
	}.Restore()
}

func NewFromConfig(selector domain.ServiceSelector, provider *domain.AWSProvider) ([]domain.ConsistentStateService, error) {
	if selector.Type != RESOURCE_TYPE {
		return nil, fmt.Errorf("Unable to create ECSService object from selector of type %s.", selector.Type)
	}

	objs := []domain.ConsistentStateService{}
	var err error

	err = domain.ValidateServiceSelector(selector)
	if err != nil {
		return nil, err
	}

	if selector.Filter != "" {
		var cluster, service string
		props := strings.Split(selector.Filter, ";")
		for _, prop := range props {
			tokens := strings.Split(prop, "=")
			key := tokens[0]
			value := tokens[1]

			if key == "cluster" {
				cluster = value
			} else if key == "service" {
				service = value
			} else {
				return nil, fmt.Errorf("Unrecognized key %s for type %s", key, RESOURCE_TYPE)
			}
		}

		objs = []domain.ConsistentStateService{
			ECSService{
				Provider:    provider,
				ClusterArn:  cluster,
				ServiceName: service,
			},
		}
	} else {
		client := ecs.NewFromConfig(provider.GetConnection())
		clusters, err := searchAllClusters(client, selector.Tags)
		if err != nil {
			return nil, err
		}

		for cluster, services := range clusters {
			for _, service := range services {
				objs = append(objs, ECSService{
					Provider:    provider,
					ClusterArn:  cluster,
					ServiceName: service,
				})
			}
		}
	}

	return objs, nil
}

func searchAllClusters(client *ecs.Client, tags []domain.AWSTag) (map[string][]string, error) {
	allClusters := map[string][]string{}

	paginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{})
	for paginator.HasMorePages() {
		response, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, cluster := range response.ClusterArns {
			serviceArns, err := filterECSServicesByTag(client, cluster, tags)
			if err != nil {
				return nil, err
			}

			allClusters[cluster] = serviceArns
		}
	}

	return allClusters, nil
}

func filterECSServicesByTag(client *ecs.Client, cluster string, tags []domain.AWSTag) ([]string, error) {
	serviceArns := []string{}

	paginator := ecs.NewListServicesPaginator(client, &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	})

	for paginator.HasMorePages() {
		response, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, arn := range response.ServiceArns {
			service, err := client.ListTagsForResource(context.TODO(), &ecs.ListTagsForResourceInput{
				ResourceArn: aws.String(arn),
			})
			if err != nil {
				return nil, err
			}

			allMatch := len(service.Tags) >= len(tags)
			for _, filterTag := range tags {
				match := false
				for _, resourceTag := range service.Tags {
					if filterTag.Name == *resourceTag.Key && filterTag.Value == *resourceTag.Value {
						match = true
					}
				}
				allMatch = allMatch && match
			}
			if allMatch {
				serviceArns = append(serviceArns, arn)
			}
		}
	}

	return serviceArns, nil
}
