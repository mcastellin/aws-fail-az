package asg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/awsutils"
	"github.com/mcastellin/aws-fail-az/state"
	"golang.org/x/exp/slices"
)

// The resource key to use for storing state of autoscaling groups
const RESOURCE_TYPE string = "ec2-asg"

func NewFromConfig(selector domain.ServiceSelector, provider *domain.AWSProvider) (*AutoscalingGroup, error) {

	if selector.Type != RESOURCE_TYPE {
		return nil, fmt.Errorf("Unable to create AutoscalingGroup object from selector of type %s.", selector.Type)
	}

	var asgName string
	tokens := strings.Split(selector.Filter, "=")
	key := tokens[0]
	value := tokens[1]

	if key == "name" {
		asgName = value
	} else {
		return nil, fmt.Errorf("Unrecognized key %s for type %s", key, RESOURCE_TYPE)
	}

	return &AutoscalingGroup{
		Provider:             provider,
		AutoScalingGroupName: asgName,
	}, nil
}

type AutoscalingGroup struct {
	Provider             *domain.AWSProvider
	AutoScalingGroupName string
}

type AutoScalingGroupState struct {
	AutoScalingGroupName string   `json:"asgName"`
	Subnets              []string `json:"subnets"`
}

func (asg AutoscalingGroup) Check() (bool, error) {
	isValid := true

	client := autoscaling.NewFromConfig(asg.Provider.GetConnection())

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asg.AutoScalingGroupName},
	}

	describeAsgOutput, err := client.DescribeAutoScalingGroups(context.TODO(), input)
	if err != nil {
		return false, err
	}

	asgObj := describeAsgOutput.AutoScalingGroups[0]
	if int(*asgObj.DesiredCapacity) > len(asgObj.Instances) {
		return false, fmt.Errorf("Desired instance capacity for AutoscalingGroup %s is not met. Desired %d, found %d.",
			asg.AutoScalingGroupName, *asgObj.DesiredCapacity, len(asgObj.Instances))
	}

	for _, instance := range asgObj.Instances {
		if *instance.HealthStatus != "Healthy" {
			return false, fmt.Errorf("Invalid health status of instance %s for AutoscalingGroup %s. Found %s.",
				*instance.InstanceId, asg.AutoScalingGroupName, *instance.HealthStatus)
		}
	}

	return isValid, nil
}

func (asg AutoscalingGroup) Save(stateManager *state.StateManager) error {

	client := autoscaling.NewFromConfig(asg.Provider.GetConnection())

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asg.AutoScalingGroupName},
	}

	describeAsgOutput, err := client.DescribeAutoScalingGroups(context.TODO(), input)
	if err != nil {
		return err
	}

	asgObj := describeAsgOutput.AutoScalingGroups[0]
	subnets := strings.Split(*asgObj.VPCZoneIdentifier, ",")

	state := &AutoScalingGroupState{
		AutoScalingGroupName: *asgObj.AutoScalingGroupName,
		Subnets:              subnets,
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Println("Error while marshalling autoscaling group state")
		return err
	}

	err = stateManager.Save(RESOURCE_TYPE, *asgObj.AutoScalingGroupName, data)
	if err != nil {
		return err
	}

	return nil
}

func (asg AutoscalingGroup) Fail(azs []string) error {
	ec2Client := ec2.NewFromConfig(asg.Provider.GetConnection())
	client := autoscaling.NewFromConfig(asg.Provider.GetConnection())

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asg.AutoScalingGroupName},
	}

	describeAsgOutput, err := client.DescribeAutoScalingGroups(context.TODO(), input)
	if err != nil {
		return err
	}

	asgObj := describeAsgOutput.AutoScalingGroups[0]
	subnets := strings.Split(*asgObj.VPCZoneIdentifier, ",")

	newSubnets, err := awsutils.FilterSubnetsNotInAzs(ec2Client, subnets, azs)

	updateAsgInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: &asg.AutoScalingGroupName,
		VPCZoneIdentifier:    aws.String(strings.Join(newSubnets, ",")),
	}

	_, err = client.UpdateAutoScalingGroup(context.TODO(), updateAsgInput)
	if err != nil {
		return err
	}

	fmt.Println(*asgObj.Instances[0].AvailabilityZone)
	instancesToTerminate := []string{}
	for _, instance := range asgObj.Instances {
		if slices.Contains(azs, *instance.AvailabilityZone) {
			instancesToTerminate = append(instancesToTerminate, *instance.InstanceId)
		}
	}
	if len(instancesToTerminate) > 0 {
		log.Printf("Terminating instances %s for autoscaling group %s that belonged to remove subnets.",
			instancesToTerminate, asg.AutoScalingGroupName)

		terminateInstancesInput := &ec2.TerminateInstancesInput{
			InstanceIds: instancesToTerminate,
		}
		_, err = ec2Client.TerminateInstances(context.TODO(), terminateInstancesInput)
		if err != nil {
			return err
		}
	}

	return nil
}
func (asg AutoscalingGroup) Restore(stateData []byte) error {
	var state AutoScalingGroupState
	err := json.Unmarshal(stateData, &state)
	if err != nil {
		return err
	}
	log.Printf("Restoring AZs for autoscaling group %s", state.AutoScalingGroupName)
	client := autoscaling.NewFromConfig(asg.Provider.GetConnection())
	updateAsgInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(state.AutoScalingGroupName),
		VPCZoneIdentifier:    aws.String(strings.Join(state.Subnets, ",")),
	}

	_, err = client.UpdateAutoScalingGroup(context.TODO(), updateAsgInput)
	if err != nil {
		return err
	}
	return nil
}
