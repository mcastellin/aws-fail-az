package ecs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type ECSService struct {
	Client      *ecs.Client
	ClusterArn  string
	ServiceName string
}

func (svc ECSService) Validate() (bool, error) {
	isValid := true

	result, err := serviceExists(svc.Client, svc.ClusterArn, svc.ServiceName)
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

func serviceExists(client *ecs.Client, clusterArn string, serviceName string) (bool, error) {

	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterArn),
		Services: []string{*aws.String(serviceName)},
	}

	_, err := client.DescribeServices(context.TODO(), input)

	if err != nil {
		if t := new(types.ResourceNotFoundException); errors.As(err, &t) {
			return false, nil
		} else if t := new(types.ClusterNotFoundException); errors.As(err, &t) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
