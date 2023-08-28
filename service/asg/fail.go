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
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/service/awsutils"
	"github.com/mcastellin/aws-fail-az/state"
	"golang.org/x/exp/slices"
)

// The resource key to use for storing state of autoscaling groups
const RESOURCE_TYPE string = "auto-scaling-group"

type AutoScalingGroupState struct {
	AutoScalingGroupName string   `json:"asgName"`
	Subnets              []string `json:"subnets"`
}

type AutoScalingGroup struct {
	Provider             awsapis.AWSProvider
	AutoScalingGroupName string

	stateSubnets []string
}

func (asg AutoScalingGroup) Check() (bool, error) {
	isValid := true

	log.Printf("%s name=%s: checking resource state before failure simulation",
		RESOURCE_TYPE, asg.AutoScalingGroupName)

	api := asg.Provider.NewAutoScalingApi()

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asg.AutoScalingGroupName},
	}

	describeAsgOutput, err := api.DescribeAutoScalingGroups(context.TODO(), input)
	if err != nil {
		return false, err
	}

	asgObj := describeAsgOutput.AutoScalingGroups[0]
	if int(*asgObj.DesiredCapacity) > len(asgObj.Instances) {
		return false, fmt.Errorf("Desired instance capacity for AutoScalingGroup %s is not met. Desired %d, found %d.",
			asg.AutoScalingGroupName, *asgObj.DesiredCapacity, len(asgObj.Instances))
	}

	for _, instance := range asgObj.Instances {
		if *instance.HealthStatus != "Healthy" {
			return false, fmt.Errorf("Invalid health status of instance %s for AutoScalingGroup %s. Found %s.",
				*instance.InstanceId, asg.AutoScalingGroupName, *instance.HealthStatus)
		}
	}

	return isValid, nil
}

func (asg AutoScalingGroup) Save(stateManager state.StateManager) error {

	api := asg.Provider.NewAutoScalingApi()

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asg.AutoScalingGroupName},
	}

	describeAsgOutput, err := api.DescribeAutoScalingGroups(context.TODO(), input)
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

func (asg AutoScalingGroup) Fail(azs []string) error {
	ec2Api := asg.Provider.NewEc2Api()
	api := asg.Provider.NewAutoScalingApi()

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asg.AutoScalingGroupName},
	}

	describeAsgOutput, err := api.DescribeAutoScalingGroups(context.TODO(), input)
	if err != nil {
		return err
	}

	asgObj := describeAsgOutput.AutoScalingGroups[0]
	subnets := strings.Split(*asgObj.VPCZoneIdentifier, ",")

	newSubnets, err := awsutils.FilterSubnetsNotInAzs(ec2Api, subnets, azs)

	log.Printf("%s name=%s: failing AZs %s for autoscaling group",
		RESOURCE_TYPE, asg.AutoScalingGroupName, azs)

	updateAsgInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.AutoScalingGroupName),
		VPCZoneIdentifier:    aws.String(strings.Join(newSubnets, ",")),
	}

	_, err = api.UpdateAutoScalingGroup(context.TODO(), updateAsgInput)
	if err != nil {
		return err
	}

	instancesToTerminate := []string{}
	for _, instance := range asgObj.Instances {
		if slices.Contains(azs, *instance.AvailabilityZone) {
			instancesToTerminate = append(instancesToTerminate, *instance.InstanceId)
		}
	}
	if len(instancesToTerminate) > 0 {
		log.Printf("%s name=%s: terminating instances %s that belonged to removed subnets",
			RESOURCE_TYPE, asg.AutoScalingGroupName, instancesToTerminate)

		terminateInstancesInput := &ec2.TerminateInstancesInput{
			InstanceIds: instancesToTerminate,
		}
		_, err = ec2Api.TerminateInstances(context.TODO(), terminateInstancesInput)
		if err != nil {
			return err
		}
	}

	return nil
}
func (asg AutoScalingGroup) Restore() error {
	log.Printf("%s name=%s: restoring AZs for autoscaling group", RESOURCE_TYPE, asg.AutoScalingGroupName)

	api := asg.Provider.NewAutoScalingApi()
	updateAsgInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.AutoScalingGroupName),
		VPCZoneIdentifier:    aws.String(strings.Join(asg.stateSubnets, ",")),
	}

	_, err := api.UpdateAutoScalingGroup(context.TODO(), updateAsgInput)
	if err != nil {
		return err
	}
	return nil
}
