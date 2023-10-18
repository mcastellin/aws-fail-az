package cmd

import (
	"fmt"
	"log"

	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/service"
	"github.com/mcastellin/aws-fail-az/state"
)

type ListCommand struct {
	Provider  awsapis.AWSProvider
	Namespace string
}

type AutoScalingGroupState struct {
	AutoScalingGroupName string   `json:"asgName"`
	Subnets              []string `json:"subnets"`
}

func (cmd *ListCommand) Run() error {
	stateManager, err := state.NewStateManager(cmd.Provider, cmd.Namespace)
	if err != nil {
		log.Print("Failed to create AWS state manager")
		return err
	}

	if err := stateManager.Initialize(); err != nil {
		return err
	}

	states, err := stateManager.QueryStates(&state.QueryStatesInput{})
	if err != nil {
		return err
	}

	ns := cmd.Namespace
	if len(ns) == 0 {
		ns = "default"
	}

	if len(states) == 0 {
		fmt.Printf("No attacked resources found for namespace '%s':\n", ns)
		return nil
	}

	fmt.Printf("Attacked resources for namespace '%s':\n", ns)
	faultTypes := service.InitServiceFaults()
	for _, s := range states {
		description, err := faultTypes.DescribeState(s)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println(description)
		}
	}

	return nil
}
