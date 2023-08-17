package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/mcastellin/aws-fail-az/domain"
)

type ECSService struct {
	Provider    *domain.AWSProvider
	ClusterArn  string
	ServiceName string
}

func (svc ECSService) Check() (bool, error) {
	isValid := true

	client := ecs.NewFromConfig(svc.Provider.GetConnection())

	result, err := serviceStable(client, svc.ClusterArn, svc.ServiceName)
	if err != nil {
		return false, nil
	} else {
		isValid = isValid && result
	}

	return isValid, nil
}

func (svc ECSService) Save() error {
	ecsClient := ecs.NewFromConfig(svc.Provider.GetConnection())

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(svc.ClusterArn),
		Services: []string{*aws.String(svc.ServiceName)},
	}

	describeOutput, err := ecsClient.DescribeServices(context.TODO(), input)
	if err != nil {
		return err
	}

	service := describeOutput.Services[0]
	subnets := service.NetworkConfiguration.AwsvpcConfiguration.Subnets

	log.Println(subnets)
	//TODO

	return nil
}

func (svc ECSService) Fail() error {
	ec2Client := ec2.NewFromConfig(svc.Provider.GetConnection())
	ecsClient := ecs.NewFromConfig(svc.Provider.GetConnection())

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(svc.ClusterArn),
		Services: []string{*aws.String(svc.ServiceName)},
	}

	describeOutput, err := ecsClient.DescribeServices(context.TODO(), input)
	if err != nil {
		return err
	}

	service := describeOutput.Services[0]
	subnets := service.NetworkConfiguration.AwsvpcConfiguration.Subnets

	newSubnets, err := filterSubnetsNotInAzs(ec2Client, subnets, []string{"us-east-1b"})
	if err != nil {
		log.Printf("Error while filtering subnets by AZs: %v", err)
		return err
	}

	if len(newSubnets) == 0 {
		return fmt.Errorf("AZ failure for service %s would remove all available subnets. Service failure will now stop.", svc.ServiceName)
	}

	updatedNetworkConfig := service.NetworkConfiguration
	updatedNetworkConfig.AwsvpcConfiguration.Subnets = newSubnets

	updateServiceInput := &ecs.UpdateServiceInput{
		Cluster:              aws.String(svc.ClusterArn),
		Service:              aws.String(svc.ServiceName),
		TaskDefinition:       service.TaskDefinition,
		NetworkConfiguration: updatedNetworkConfig,
	}

	_, err = ecsClient.UpdateService(context.TODO(), updateServiceInput)
	if err != nil {
		return err
	}

	err = stopTasksInRemovedSubnets(ecsClient, svc.ClusterArn, svc.ServiceName, newSubnets)
	if err != nil {
		return err
	}

	return nil
}

func (svc ECSService) Restore() error {
	return nil
}
