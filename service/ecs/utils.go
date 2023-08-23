package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"golang.org/x/exp/slices"
)

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
						RESOURCE_TYPE, cluster, service, *task.TaskArn)
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
