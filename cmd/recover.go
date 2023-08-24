package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/state"
)

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

	states, err := stateManager.ReadStates(&state.QueryStatesInput{})
	if err != nil {
		log.Panic(err)
	}
	for _, s := range states {
		if s.ResourceType == ecs.RESOURCE_TYPE {
			err = ecs.RestoreFromState(s.State, &provider)
		} else if s.ResourceType == asg.RESOURCE_TYPE {
			err = asg.RestoreFromState(s.State, &provider)
		} else {
			err = fmt.Errorf("Unknown resource of type %s found in state with key %s. Object will be ignored.\n",
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
