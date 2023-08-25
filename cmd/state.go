package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

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
	stateData string) {

	var statePayload []byte
	var err error
	if readFromStdin {
		statePayload, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Panic(err)
		}
	} else {
		statePayload = []byte(stateData)
	}

	if len(statePayload) == 0 {
		log.Fatal("No data was provided to store in state. Exiting.")
	}

	stateManager := getManager(namespace)
	stateManager.Initialize()

	err = stateManager.Save(resourceType, resourceKey, statePayload)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadStates(namespace string, resourceType string, resourceKey string) {

	// Discard logging to facilitate output parsing
	log.SetOutput(io.Discard)

	stateManager := getManager(namespace)
	stateManager.Initialize()

	states, err := stateManager.ReadStates(&state.QueryStatesInput{
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
		stateJson, err := json.Marshal(stateData)
		if err != nil {
			fmt.Println("Error unmarshalling state object. Exiting.")
		}
		fmt.Println(string(stateJson))
	} else {
		fmt.Println("[]")
	}
}

func DeleteState(namespace string, resourceType string, resourceKey string) {

	stateManager := getManager(namespace)
	stateManager.Initialize()

	result, err := stateManager.GetState(resourceType, resourceKey)
	if err != nil {
		log.Fatal(err)
	}

	err = stateManager.RemoveState(*result)
	if err != nil {
		log.Fatalf("Error removing state object with key %s", result.Key)
	} else {
		log.Printf("State with key %s removed successfully", result.Key)
	}
}
