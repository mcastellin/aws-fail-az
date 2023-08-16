package ecs

import (
	"context"
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
	log.Println(subnets)

	filterSubnetsByAz(ec2Client, subnets, []string{"us-east-1a"})

	return nil
}

func filterSubnetsByAz(client *ec2.Client, subnetIds []string, azs []string) ([]string, error) {
	//TODO

	return []string{}, nil
}

func (svc ECSService) Restore() error {
	return nil
}
