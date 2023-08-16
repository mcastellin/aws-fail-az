package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/mcastellin/aws-fail-az/domain"
	ecsSvc "github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/state"
)

type BaseFilter struct {
	tags []string
}
type EcsFilter struct {
	BaseFilter

	clusterArn  string
	serviceName string
}

func validate(svc domain.ConsistentServiceState, ch chan<- bool, wg *sync.WaitGroup) {

	defer wg.Done()
	isValid, err := svc.Validate()
	if err == nil {
		log.Println("Service is invalid")
		ch <- false

	} else {
		ch <- isValid
	}
}

func main() {

	var azs = [...]string{"us-east-1"}

	fmt.Println(azs)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)
	state := state.StateManager{
		Client: dynamodbClient,
	}

	state.Initialize()

	allServices := make([]domain.ConsistentServiceState, 0)

	ecsService := ecsSvc.ECSService{
		Client:      ecs.NewFromConfig(cfg),
		ClusterArn:  "test",
		ServiceName: "test",
	}
	allServices = append(allServices, ecsService)
	allServices = append(allServices, ecsService)

	validationResults := make(chan bool, len(allServices))

	var wg sync.WaitGroup
	for _, svc := range allServices {
		wg.Add(1)
		go validate(svc, validationResults, &wg)
	}

	wg.Wait()
	close(validationResults)

	for result := range validationResults {
		log.Println(result)
	}
}
