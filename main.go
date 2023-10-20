package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/cmd"
	"github.com/spf13/cobra"
)

// BuildVersion for this application
var BuildVersion string

var (
	awsRegion         string
	awsProfile        string
	stdin             bool
	namespace         string
	resourceType      string
	resourceKey       string
	resourceStateData string
)

var rootCmd = &cobra.Command{
	Use:   "aws-fail-az",
	Short: "aws-fail-az is an AWS utility to simulate Availability Zone failure",
}

var failCmd = &cobra.Command{
	Use:   "fail [CONFIG_FILE]",
	Short: "Start AZ failure injection based on the provided configuration from stdin",
	RunE: func(_ *cobra.Command, args []string) error {
		if !stdin && len(args) != 1 {
			return fmt.Errorf("Only one fault configuration file should be provided. Found %d.", len(args))
		} else if stdin && len(args) > 0 {
			return fmt.Errorf("Configuration files are not supported when reading from stdin. Found %d.", len(args))
		}
		configFile := ""
		if !stdin {
			configFile = args[0]
		}
		provider, err := createProvider()
		if err != nil {
			return err
		}
		op := &cmd.FailCommand{
			Provider:      provider,
			Namespace:     namespace,
			ReadFromStdin: stdin,
			ConfigFile:    configFile,
		}
		return op.Run()
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover from AZ failure and restore saved resources state",
	RunE: func(_ *cobra.Command, args []string) error {
		provider, err := createProvider()
		if err != nil {
			return err
		}
		op := &cmd.RecoverCommand{Provider: provider, Namespace: namespace}
		return op.Run()
	},
}

var stateSaveCmd = &cobra.Command{
	Use:   "state-save",
	Short: "Store a state object in Dynamodb",
	RunE: func(_ *cobra.Command, args []string) error {
		if stdin && len(resourceStateData) > 0 {
			return fmt.Errorf("State files are not supported when reading from stdin. Found %d.", len(args))
		}

		provider, err := createProvider()
		if err != nil {
			return err
		}
		op := &cmd.SaveStateCommand{
			Provider:      provider,
			Namespace:     namespace,
			ResourceType:  resourceType,
			ResourceKey:   resourceKey,
			ReadFromStdin: stdin,
			StateData:     resourceStateData,
		}
		return op.Run()
	},
}

var stateReadCmd = &cobra.Command{
	Use:   "state-read",
	Short: "Read a state object from Dynamodb",
	RunE: func(_ *cobra.Command, args []string) error {
		provider, err := createProvider()
		if err != nil {
			return err
		}
		op := &cmd.ReadStatesCommand{
			Provider:     provider,
			Namespace:    namespace,
			ResourceType: resourceType,
			ResourceKey:  resourceKey,
		}
		return op.Run()
	},
}

var stateDeleteCmd = &cobra.Command{
	Use:   "state-delete",
	Short: "Delete a state object from Dynamodb",
	RunE: func(_ *cobra.Command, args []string) error {
		provider, err := createProvider()
		if err != nil {
			return err
		}
		op := &cmd.DeleteStateCommand{
			Provider:     provider,
			Namespace:    namespace,
			ResourceType: resourceType,
			ResourceKey:  resourceKey,
		}
		return op.Run()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the attacked resources",
	RunE: func(_ *cobra.Command, args []string) error {
		provider, err := createProvider()
		if err != nil {
			return err
		}
		op := &cmd.ListCommand{
			Provider:  provider,
			Namespace: namespace,
		}
		return op.Run()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the command version",
	Run: func(_ *cobra.Command, args []string) {
		fmt.Printf("aws-fail-az v%s\n", BuildVersion)
	},
}

func createProvider() (awsapis.AWSProvider, error) {
	config.WithSharedConfigProfile("devlearnops")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(awsProfile),
		config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("Failed to load AWS configuration: %v", err)
	}

	return awsapis.NewProviderFromConfig(&cfg), nil
}

func main() {

	failCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")
	failCmd.Flags().BoolVar(&stdin, "stdin", false, "Read fail configuration from stdin.")

	recoverCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")

	stateSaveCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")
	stateSaveCmd.Flags().StringVar(&resourceType, "type", "", "The type of resource state to store")
	stateSaveCmd.Flags().StringVar(&resourceKey, "key", "", "A unique key to identify this resource")
	stateSaveCmd.Flags().StringVar(&resourceStateData, "data", "", "The payload for the resource state as a string value")
	stateSaveCmd.Flags().BoolVar(&stdin, "stdin", false, "Read resource state from stdin.")
	stateSaveCmd.MarkFlagRequired("type")
	stateSaveCmd.MarkFlagRequired("key")

	stateReadCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")
	stateReadCmd.Flags().StringVar(&resourceType, "type", "", "Filter states by resource type")
	stateReadCmd.Flags().StringVar(&resourceKey, "key", "", "Filter states by resource key")

	stateDeleteCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")
	stateDeleteCmd.Flags().StringVar(&resourceType, "type", "", "Filter states by resource type")
	stateDeleteCmd.Flags().StringVar(&resourceKey, "key", "", "Filter states by resource key")
	stateDeleteCmd.MarkFlagRequired("type")
	stateDeleteCmd.MarkFlagRequired("key")

	listCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")

	rootCmd.PersistentFlags().StringVar(&awsRegion, "region", "", "The AWS region")
	rootCmd.PersistentFlags().StringVar(&awsProfile, "profile", "", "The AWS profile")
	rootCmd.AddCommand(failCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(stateSaveCmd)
	rootCmd.AddCommand(stateReadCmd)
	rootCmd.AddCommand(stateDeleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
