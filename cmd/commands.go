package cmd

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/state"
)

func validate(svc domain.ConsistentServiceState, ch chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()
	isValid, err := svc.Check()
	if err != nil {
		ch <- false
	} else {
		ch <- isValid
	}
}

func FailCommand(namespace string, readFromStdin bool, configFile string) {

	var configContent []byte
	var err error
	if readFromStdin {
		configContent, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Panic(err)
		}
	} else {
		configContent, err = os.ReadFile(configFile)
		if err != nil {
			log.Panic(err)
		}
	}

	var faultConfig domain.FaultConfiguration
	err = json.Unmarshal(configContent, &faultConfig)
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
		Client:    dynamodbClient,
		Namespace: namespace,
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

	for _, svc := range allServices {
		err = svc.Save(stateManager)
		if err != nil {
			log.Panic(err)
		}
	}

	for _, svc := range allServices {
		err = svc.Fail(faultConfig.Azs)
		if err != nil {
			log.Panic(err)
		}
	}
}

func RecoverCommand(namespace string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}
	provider := domain.NewProviderFromConfig(&cfg)

	dynamodbClient := dynamodb.NewFromConfig(cfg)
	stateManager := &state.StateManager{
		Client:    dynamodbClient,
		Namespace: namespace,
	}

	stateManager.Initialize()

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
