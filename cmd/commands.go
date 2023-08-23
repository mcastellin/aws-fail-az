package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/state"
)

func validate(svc domain.ConsistentStateService, ch chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()
	isValid, err := svc.Check()
	if err != nil {
		log.Println(err)
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

	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager := &state.StateManagerImpl{
		Api:       awsapis.NewDynamodbApi(&provider),
		Namespace: namespace,
	}

	stateManager.Initialize()

	allServices := make([]domain.ConsistentStateService, 0)

	for _, svc := range faultConfig.Services {
		var svcConfigs []domain.ConsistentStateService
		var err error

		switch {
		case svc.Type == ecs.RESOURCE_TYPE:
			svcConfigs, err = ecs.NewFromConfig(svc, &provider)
		case svc.Type == asg.RESOURCE_TYPE:
			svcConfigs, err = asg.NewFromConfig(svc, &provider)
		default:
			err = fmt.Errorf("Could not recognize resource type %s", svc.Type)
		}
		if err != nil {
			log.Panic(err)
		}
		allServices = append(allServices, svcConfigs...)

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
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager := &state.StateManagerImpl{
		Api:       awsapis.NewDynamodbApi(&provider),
		Namespace: namespace,
	}

	stateManager.Initialize()

	states, err := stateManager.ReadStates()
	if err != nil {
		log.Panic(err)
	}
	for _, s := range states {
		if s.ResourceType == ecs.RESOURCE_TYPE {
			err = ecs.RestoreFromState(s.State, &provider)
		} else if s.ResourceType == asg.RESOURCE_TYPE {
			err = asg.RestoreFromState(s.State, &provider)
		} else {
			err = fmt.Errorf("Unknown resource of type %s found in state for key %s. Could not recover.\n",
				s.ResourceType,
				s.Key,
			)
		}

		if err != nil {
			log.Println(err)
		} else {
			stateManager.RemoveState(s)
		}
	}
}
