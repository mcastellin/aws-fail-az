package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// BuildVersion for this application
var BuildVersion string
var (
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
	Run: func(cmd *cobra.Command, args []string) {
		if !stdin && len(args) != 1 {
			log.Fatalf("Only one fault coniguration file should be provided. Found %d.", len(args))
		} else if stdin && len(args) > 0 {
			log.Fatalf("Configuration files are not supported when reading from stdin. Found %d.", len(args))
		}
		configFile := ""
		if !stdin {
			configFile = args[0]
		}
		FailCommand(namespace, stdin, configFile)
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover from AZ failure and restore saved resources state",
	Run: func(cmd *cobra.Command, args []string) {
		RecoverCommand(namespace)
	},
}

var stateSaveCmd = &cobra.Command{
	Use:   "state-save",
	Short: "Store a state object in Dynamodb",
	Run: func(cmd *cobra.Command, args []string) {
		if stdin && len(resourceStateData) > 0 {
			log.Fatalf("State files are not supported when reading from stdin. Found %d.", len(args))
		}
		SaveState(namespace, resourceType, resourceKey, stdin, resourceStateData)
	},
}

var stateReadCmd = &cobra.Command{
	Use:   "state-read",
	Short: "Read a state object from Dynamodb",
	Run: func(cmd *cobra.Command, args []string) {
		ReadStates(namespace, resourceType, resourceKey)
	},
}

var stateDeleteCmd = &cobra.Command{
	Use:   "state-delete",
	Short: "Delete a state object from Dynamodb",
	Run: func(cmd *cobra.Command, args []string) {
		DeleteState(namespace, resourceType, resourceKey)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the command version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("aws-fail-az v%s\n", BuildVersion)
	},
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

	rootCmd.AddCommand(failCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(stateSaveCmd)
	rootCmd.AddCommand(stateReadCmd)
	rootCmd.AddCommand(stateDeleteCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
