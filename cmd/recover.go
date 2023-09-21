package cmd

import (
	"fmt"
	"log"

	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/service/elbv2"
	"github.com/mcastellin/aws-fail-az/state"
)

type RecoverCommand struct {
	Provider  awsapis.AWSProvider
	Namespace string
}

func (cmd *RecoverCommand) Run() error {
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
	for _, s := range states {
		switch s.ResourceType {
		case ecs.RESOURCE_TYPE:
			err = ecs.RestoreFromState(s.State, cmd.Provider)
		case asg.RESOURCE_TYPE:
			err = asg.RestoreFromState(s.State, cmd.Provider)
		case elbv2.RESOURCE_TYPE:
			err = elbv2.RestoreFromState(s.State, cmd.Provider)
		default:
			err = fmt.Errorf("unknown resource of type %s found in state with key %s. Object will be ignored",
				s.ResourceType,
				s.Key,
			)
		}

		if err != nil {
			log.Println(err)
		} else {
			err = stateManager.RemoveState(s)
			if err != nil {
				log.Printf("Error removing state from storage: %v", err)
			}
		}
	}

	return nil
}
