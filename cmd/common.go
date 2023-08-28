package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/state"
)

func getManager(namespace string) state.StateManager {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	provider := awsapis.NewProviderFromConfig(&cfg)

	return &state.StateManagerImpl{
		Api:       provider.NewDynamodbApi(),
		Namespace: namespace,
	}
}
