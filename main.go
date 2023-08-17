package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/ecs"
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
	isValid, err := svc.Check()
	if err != nil {
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

	provider := domain.NewProviderFromConfig(&cfg)

	dynamodbClient := dynamodb.NewFromConfig(cfg)
	state := state.StateManager{
		Client: dynamodbClient,
	}

	state.Initialize()

	allServices := make([]domain.ConsistentServiceState, 0)
	ecsService := ecs.ECSService{
		Provider:    &provider,
		ClusterArn:  "tutorial-sample-app-cluster",
		ServiceName: "sample-app-back",
	}
	allServices = append(allServices, ecsService)

	//ecsService = ecs.ECSService{
	//Provider:    &provider,
	//ClusterArn:  "test",
	//ServiceName: "test",
	//}
	//allServices = append(allServices, ecsService)

	validationResults := make(chan bool, len(allServices))

	var wg sync.WaitGroup
	for _, svc := range allServices {
		wg.Add(1)
		go validate(svc, validationResults, &wg)
	}

	wg.Wait()
	close(validationResults)

	for isValid := range validationResults {
		if !isValid {
			log.Panic("One or more resources failed state checks. Panic.")
		}
	}

	err = ecsService.Save()
	if err != nil {
		log.Panic(err)
	}

	//err = ecsService.Fail()
	//if err != nil {
	//log.Panic(err)
	//}
}
