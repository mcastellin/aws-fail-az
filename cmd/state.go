package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/state"
)

type ReadStatesOutput struct {
	Namespace    string `json:"namespace"`
	ResourceType string `json:"type"`
	ResourceKey  string `json:"key"`
	State        string `json:"state"`
}

func SaveState(namespace string,
	resourceType string,
	resourceKey string,
	readFromStdin bool,
	stateData string) error {

	var statePayload []byte
	var err error
	if readFromStdin {
		statePayload, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		statePayload = []byte(stateData)
	}

	if len(statePayload) == 0 {
		return fmt.Errorf("No data was provided to store in state. Exiting.")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to load AWS configuration: %v", err)
	}
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager, err := state.NewStateManager(provider, namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}
	if err := stateManager.Initialize(); err != nil {
		return err
	}

	err = stateManager.Save(resourceType, resourceKey, statePayload)
	if err != nil {
		return err
	}

	return nil
}

func ReadStates(namespace string, resourceType string, resourceKey string) error {
	// Discard logging to facilitate output parsing
	log.SetOutput(io.Discard)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to load AWS configuration: %v", err)
	}
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager, err := state.NewStateManager(provider, namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}
	if err := stateManager.Initialize(); err != nil {
		return err
	}

	states, err := stateManager.QueryStates(&state.QueryStatesInput{
		ResourceType: resourceType,
		ResourceKey:  resourceKey,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stateData := []ReadStatesOutput{}
	for _, s := range states {
		stateData = append(stateData,
			ReadStatesOutput{
				Namespace:    s.Namespace,
				ResourceType: s.ResourceType,
				ResourceKey:  s.ResourceKey,
				State:        string(s.State),
			})
	}

	if len(states) > 0 {
		stateJSON, err := json.Marshal(stateData)
		if err != nil {
			fmt.Println("Error unmarshalling state object. Exiting.")
		}
		fmt.Println(string(stateJSON))
	} else {
		fmt.Println("[]")
	}

	return nil
}

func DeleteState(namespace string, resourceType string, resourceKey string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to load AWS configuration: %v", err)
	}
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager, err := state.NewStateManager(provider, namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}
	if err := stateManager.Initialize(); err != nil {
		return err
	}

	result, err := stateManager.GetState(resourceType, resourceKey)
	if err != nil {
		return err
	}

	err = stateManager.RemoveState(*result)
	if err != nil {
		log.Print("Error removing state object with key %s", result.Key)
		return err
	}
	log.Printf("State with key %s removed successfully", result.Key)

	return nil
}
