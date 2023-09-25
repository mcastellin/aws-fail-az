package cmd

import (
	"log"

	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/service"
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

	faultTypes := service.InitServiceFaults()
	for _, s := range states {
		err := faultTypes.RestoreFromState(s, cmd.Provider)
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
