package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/state"
)

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
		Api:       provider.NewDynamodbApi(),
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

	log.Println("INFO: Checking resources state is stable before AZ failure.")
	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(15*time.Minute))
	defer cancel()

	err = checkResourceStates(ctx, allServices)
	if err != nil {
		log.Println(err)
		log.Fatal("Exiting.")
	}

	log.Println("INFO: Saving resources' states in state table.")
	for _, svc := range allServices {
		err = svc.Save(stateManager)
		if err != nil {
			log.Panic(err)
		}
	}

	log.Println("INFO: Failing configured AZs.")
	for _, svc := range allServices {
		err = svc.Fail(faultConfig.Azs)
		if err != nil {
			log.Panic(err)
		}
	}
}

func checkResourceStates(ctx context.Context, resources []domain.ConsistentStateService) error {
	checkResults := make(chan bool, len(resources))

	wg := new(sync.WaitGroup)
	for _, resource := range resources {
		wg.Add(1)
		go func(resource domain.ConsistentStateService) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				checkResults <- false
			default:
				isValid, err := resource.Check()
				if err != nil {
					log.Println(err)
					isValid = false
				}
				checkResults <- isValid
			}
		}(resource)
	}

	wg.Wait()
	close(checkResults)

	validCount := 0
	for isValid := range checkResults {
		if isValid {
			validCount++
		}
	}
	if validCount < len(resources) {
		return fmt.Errorf("ERROR: One or more resources failed state checks")
	}
	return nil
}
