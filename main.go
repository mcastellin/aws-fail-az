package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/state"
)

const faultConfiguration string = `
{
  "azs": [
    "us-east-1a"
  ],
  "services": [
    {
      "type": "ecs-service",
      "filter": "cluster=tutorial-sample-app-cluster;service=sample-app-back",
      "tags": [
        {
          "Name": "Environment",
          "Value": "live"
        }
      ]
    }
  ]
}
`

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

	var faultConfig domain.FaultConfiguration
	err := json.Unmarshal([]byte(faultConfiguration), &faultConfig)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Failing availability zones %s", faultConfig.Azs)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	provider := domain.NewProviderFromConfig(&cfg)

	dynamodbClient := dynamodb.NewFromConfig(cfg)
	stateManager := &state.StateManager{
		Client: dynamodbClient,
	}

	stateManager.Initialize()

	allServices := make([]domain.ConsistentServiceState, 0)

	for _, svc := range faultConfig.Services {
		if svc.Type == ecs.RESOURCE_TYPE {
			svcConfig, err := ecs.NewFromConfig(svc, &provider)
			if err != nil {
				log.Panic(err)
			} else {
				allServices = append(allServices, svcConfig)
			}
		}
	}

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

	err = allServices[0].Save(stateManager)
	if err != nil {
		log.Panic(err)
	}

	err = allServices[0].Fail(faultConfig.Azs)
	if err != nil {
		log.Panic(err)
	}

	time.Sleep(30 * time.Second)

	states, err := stateManager.ReadStates()
	if err != nil {
		log.Panic(err)
	}
	for _, s := range states {
		err = ecs.ECSService{Provider: &provider}.Restore(s.State)
		if err != nil {
			log.Println(err)
		} else {
			stateManager.RemoveState(s)
		}
	}
}
