package ecs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
)

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
