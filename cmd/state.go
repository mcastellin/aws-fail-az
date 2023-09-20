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

type SaveState struct {
	Namespace     string
	ResourceType  string
	ResourceKey   string
	ReadFromStdin bool
	StateData     string
}

func (cmd *SaveState) Run() error {

	var statePayload []byte
	var err error
	if cmd.ReadFromStdin {
		statePayload, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		statePayload = []byte(cmd.StateData)
	}

	if len(statePayload) == 0 {
		return fmt.Errorf("No data was provided to store in state. Exiting.")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to load AWS configuration: %v", err)
	}
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager, err := state.NewStateManager(provider, cmd.Namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}
	if err := stateManager.Initialize(); err != nil {
		return err
	}

	err = stateManager.Save(cmd.ResourceType, cmd.ResourceKey, statePayload)
	if err != nil {
		return err
	}

	return nil
}

type ReadStates struct {
	Namespace    string
	ResourceType string
	ResourceKey  string
}

func (cmd *ReadStates) Run() error {
	// Discard logging to facilitate output parsing
	log.SetOutput(io.Discard)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to load AWS configuration: %v", err)
	}
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager, err := state.NewStateManager(provider, cmd.Namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}
	if err := stateManager.Initialize(); err != nil {
		return err
	}

	states, err := stateManager.QueryStates(&state.QueryStatesInput{
		ResourceType: cmd.ResourceType,
		ResourceKey:  cmd.ResourceKey,
	})
	if err != nil {
		return err
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
			return err
		}
		fmt.Println(string(stateJSON))
	} else {
		fmt.Println("[]")
	}

	return nil
}

type DeleteState struct {
	Namespace    string
	ResourceType string
	ResourceKey  string
}

func (cmd *DeleteState) Run() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Failed to load AWS configuration: %v", err)
	}
	provider := awsapis.NewProviderFromConfig(&cfg)

	stateManager, err := state.NewStateManager(provider, cmd.Namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}
	if err := stateManager.Initialize(); err != nil {
		return err
	}

	result, err := stateManager.GetState(cmd.ResourceType, cmd.ResourceKey)
	if err != nil {
		return err
	}

	err = stateManager.RemoveState(*result)
	if err != nil {
		log.Printf("Error removing state object with key %s", result.Key)
		return err
	}
	log.Printf("State with key %s removed successfully", result.Key)

	return nil
}
