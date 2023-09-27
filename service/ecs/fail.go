package ecs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/awsutils"
	"github.com/mcastellin/aws-fail-az/state"
	"golang.org/x/exp/slices"
)

// A struct to represent an ECS service resource
type ECSService struct {
	Provider    awsapis.AWSProvider
	ClusterArn  string
	ServiceName string

	stateSubnets []string
}

// A struct to represent the current state of an ECS service before
// AZ failure is applied
type ECSServiceState struct {
	ServiceName string   `json:"service"`
	ClusterArn  string   `json:"cluster"`
	Subnets     []string `json:"subnets"`
}

func (svc *ECSService) Check() (bool, error) {
	isValid := true

	log.Printf("%s cluster=%s,name=%s: checking resource state before failure simulation",
		domain.ResourceTypeEcsService, svc.ClusterArn, svc.ServiceName)

	api := svc.Provider.NewEcsApi()

	result, err := serviceStable(api, svc.ClusterArn, svc.ServiceName)
	if err != nil {
		return false, err
	}

	return isValid && result, nil
}

func (svc *ECSService) Save(stateManager state.StateManager) error {
	api := svc.Provider.NewEcsApi()

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(svc.ClusterArn),
		Services: []string{*aws.String(svc.ServiceName)},
	}

	describeOutput, err := api.DescribeServices(context.TODO(), input)
	if err != nil {
		return err
	}

	service := describeOutput.Services[0]
	subnets := service.NetworkConfiguration.AwsvpcConfiguration.Subnets

	state := &ECSServiceState{
		ClusterArn:  svc.ClusterArn,
		ServiceName: svc.ServiceName,
		Subnets:     subnets,
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Println("Error while marshalling service state")
		return err
	}

	resourceKey := fmt.Sprintf("%s-%s", svc.ClusterArn, svc.ServiceName)
	err = stateManager.Save(domain.ResourceTypeEcsService, resourceKey, data)
	if err != nil {
		return err
	}

	return nil
}

func (svc *ECSService) Fail(azs []string) error {
	ec2Api := svc.Provider.NewEc2Api()
	ecsApi := svc.Provider.NewEcsApi()

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(svc.ClusterArn),
		Services: []string{*aws.String(svc.ServiceName)},
	}

	describeOutput, err := ecsApi.DescribeServices(context.TODO(), input)
	if err != nil {
		return err
	}

	service := describeOutput.Services[0]
	subnets := service.NetworkConfiguration.AwsvpcConfiguration.Subnets

	newSubnets, err := awsutils.FilterSubnetsNotInAzs(ec2Api, subnets, azs)
	if err != nil {
		log.Printf("Error while filtering subnets by AZs: %v", err)
		return err
	}

	if len(newSubnets) == 0 {
		return fmt.Errorf("AZ failure for service %s would remove all available subnets. Service failure will now stop", svc.ServiceName)
	}

	log.Printf("%s cluster=%s,name=%s: failing AZs %s for ecs-service",
		domain.ResourceTypeEcsService, svc.ClusterArn, svc.ServiceName, azs)

	updatedNetworkConfig := service.NetworkConfiguration
	updatedNetworkConfig.AwsvpcConfiguration.Subnets = newSubnets

	updateServiceInput := &ecs.UpdateServiceInput{
		Cluster:              aws.String(svc.ClusterArn),
		Service:              aws.String(svc.ServiceName),
		TaskDefinition:       service.TaskDefinition,
		NetworkConfiguration: updatedNetworkConfig,
	}

	_, err = ecsApi.UpdateService(context.TODO(), updateServiceInput)
	if err != nil {
		return err
	}

	err = stopTasksInRemovedSubnets(ecsApi, svc.ClusterArn, svc.ServiceName, newSubnets)
	if err != nil {
		return err
	}

	return nil
}

func (svc *ECSService) Restore() error {
	log.Printf("%s cluster=%s,name=%s: restoring AZs for ecs-service",
		domain.ResourceTypeEcsService, svc.ClusterArn, svc.ServiceName)

	api := svc.Provider.NewEcsApi()

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(svc.ClusterArn),
		Services: []string{*aws.String(svc.ServiceName)},
	}

	describeOutput, err := api.DescribeServices(context.TODO(), input)
	if err != nil {
		return err
	}

	service := describeOutput.Services[0]

	updatedNetworkConfig := service.NetworkConfiguration
	updatedNetworkConfig.AwsvpcConfiguration.Subnets = svc.stateSubnets

	updateServiceInput := &ecs.UpdateServiceInput{
		Cluster:              aws.String(svc.ClusterArn),
		Service:              aws.String(svc.ServiceName),
		TaskDefinition:       service.TaskDefinition,
		NetworkConfiguration: updatedNetworkConfig,
	}

	_, err = api.UpdateService(context.TODO(), updateServiceInput)
	if err != nil {
		return err
	}
	return nil
}

// Search and terminate tasks that have an attachment to subnets that have been eliminated from
// the network configuration
func stopTasksInRemovedSubnets(api awsapis.EcsApi, cluster string, service string, validSubnets []string) error {
	paginator := api.NewListTasksPaginator(&ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	})

	for paginator.HasMorePages() {
		listTasksOutput, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		describeTasksOutput, err := api.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
			Cluster: aws.String(cluster),
			Tasks:   listTasksOutput.TaskArns,
		})
		if err != nil {
			return err
		}

		for _, task := range describeTasksOutput.Tasks {
			taskSubnets := getTaskSubnets(task)
			for _, sub := range taskSubnets {
				if !slices.Contains(validSubnets, sub) {
					stopTaskInput := &ecs.StopTaskInput{
						Cluster: aws.String(cluster),
						Task:    task.TaskArn,
						Reason:  aws.String("AZ failure simulation. Task belonged to removed subnet."),
					}
					_, err = api.StopTask(context.TODO(), stopTaskInput)
					if err != nil {
						return err
					}
					log.Printf("%s cluster=%s,name=%s: terminating task %s running in removed subnets.",
						domain.ResourceTypeEcsService, cluster, service, *task.TaskArn)
				}
			}
		}
	}

	return nil
}

// Returns the list of subnets attached to an ECS task
func getTaskSubnets(task ecsTypes.Task) []string {
	subnets := []string{}
	for _, attachment := range task.Attachments {
		for _, detail := range attachment.Details {
			if *detail.Name == "subnetId" {
				subnets = append(subnets, *detail.Value)
			}
		}
	}
	return subnets
}

// Verify if it's safe to fail service availability zones
// In order to avoid compromising already unstable services, this method verifies that
// the service exists and has currently reached a stable state.
func serviceStable(api awsapis.EcsApi, clusterArn string, serviceName string) (bool, error) {

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterArn),
		Services: []string{*aws.String(serviceName)},
	}

	describeOutput, err := api.DescribeServices(context.TODO(), input)

	if err != nil {
		if t := new(types.ResourceNotFoundException); errors.As(err, &t) {
			return false, nil
		} else if t := new(types.ClusterNotFoundException); errors.As(err, &t) {
			return false, nil
		}
		return false, err
	}

	if len(describeOutput.Services) == 0 {
		return false, nil
	}
	svc := describeOutput.Services[0]

	serviceStable := *svc.Status == "ACTIVE" &&
		svc.DesiredCount == svc.RunningCount

	return serviceStable, nil
}
