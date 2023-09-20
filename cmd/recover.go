package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/service/elbv2"
	"github.com/mcastellin/aws-fail-az/state"
)

type RecoverCommand struct {
	Namespace string
}

func (cmd *RecoverCommand) Run() error {
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

	states, err := stateManager.QueryStates(&state.QueryStatesInput{})
	if err != nil {
		return err
	}
	for _, s := range states {
		switch s.ResourceType {
		case ecs.RESOURCE_TYPE:
			err = ecs.RestoreFromState(s.State, provider)
		case asg.RESOURCE_TYPE:
			err = asg.RestoreFromState(s.State, provider)
		case elbv2.RESOURCE_TYPE:
			err = elbv2.RestoreFromState(s.State, provider)
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
