package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"golang.org/x/exp/slices"
)

func stopTasksInRemovedSubnets(client *ecs.Client, cluster string, service string, validSubnets []string) error {
	listTasksInput := &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	}
	listTasksOutput, err := client.ListTasks(context.TODO(), listTasksInput)
	if err != nil {
		return err
	}

	describeTasksInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   listTasksOutput.TaskArns,
	}
	describeTasksOutput, err := client.DescribeTasks(context.TODO(), describeTasksInput)
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
				_, err = client.StopTask(context.TODO(), stopTaskInput)
				if err != nil {
					return err
				}
				log.Printf("Stopped task %s for service %s az-failure", *task.TaskArn, service)
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
