package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/service/elbv2"
	"github.com/mcastellin/aws-fail-az/state"
)

type FailCommand struct {
	Provider      awsapis.AWSProvider
	Namespace     string
	ReadFromStdin bool
	ConfigFile    string
}

func (cmd *FailCommand) Run() error {

	var configContent []byte
	var err error
	if cmd.ReadFromStdin {
		configContent, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		configContent, err = os.ReadFile(cmd.ConfigFile)
		if err != nil {
			return err
		}
	}

	var faultConfig domain.FaultConfiguration
	err = json.Unmarshal(configContent, &faultConfig)
	if err != nil {
		return err
	}

	log.Printf("Failing availability zones %s", faultConfig.Azs)

	stateManager, err := state.NewStateManager(cmd.Provider, cmd.Namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}

	if err := stateManager.Initialize(); err != nil {
		return err
	}

	allServices := make([]domain.ConsistentStateResource, 0)

	for _, target := range faultConfig.Targets {
		var targetConfigs []domain.ConsistentStateResource
		var err error

		switch {
		case target.Type == ecs.RESOURCE_TYPE:
			targetConfigs, err = ecs.NewFromConfig(target, cmd.Provider)
		case target.Type == asg.RESOURCE_TYPE:
			targetConfigs, err = asg.NewFromConfig(target, cmd.Provider)
		case target.Type == elbv2.RESOURCE_TYPE:
			targetConfigs, err = elbv2.NewFromConfig(target, cmd.Provider)
		default:
			err = fmt.Errorf("Could not recognize resource type %s", target.Type)
		}
		if err != nil {
			return err
		}
		allServices = append(allServices, targetConfigs...)

	}

	log.Println("INFO: Checking resources state is stable before AZ failure.")
	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(15*time.Minute))
	defer cancel()

	err = checkResourceStates(ctx, allServices)
	if err != nil {
		return err
	}

	log.Println("INFO: Saving resources' states in state table.")
	for _, svc := range allServices {
		err = svc.Save(stateManager)
		if err != nil {
			return err
		}
	}

	log.Println("INFO: Failing configured AZs.")
	for _, svc := range allServices {
		err = svc.Fail(faultConfig.Azs)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkResourceStates(ctx context.Context, resources []domain.ConsistentStateResource) error {
	checkResults := make(chan bool, len(resources))

	wg := new(sync.WaitGroup)
	for _, resource := range resources {
		wg.Add(1)
		go func(resource domain.ConsistentStateResource) {
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
