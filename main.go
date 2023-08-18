package main

import (
	"fmt"
	"log"

	_cmd "github.com/mcastellin/aws-fail-az/cmd"
	"github.com/spf13/cobra"
)

var version string = "0.0.1"
var (
	namespace string
	stdin     bool
)

var rootCmd = &cobra.Command{
	Use:   "aws-fail-az",
	Short: "aws-fail-az is an AWS utility to simulate Availability Zone failure",
}

var failCmd = &cobra.Command{
	Use:   "fail",
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
		_cmd.FailCommand(namespace, stdin, configFile)
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover from AZ failure and restore saved resources state",
	Run: func(cmd *cobra.Command, args []string) {
		_cmd.RecoverCommand(namespace)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the command version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("aws-fail-az v%s\n", version)
	},
}

func main() {

	failCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")
	failCmd.Flags().BoolVar(&stdin, "stdin", false, "Read fail configuration from stdin.")

	recoverCmd.Flags().StringVar(&namespace, "ns", "", "The namespace assigned to this operation. Used to uniquely identify resources state for recovery.")

	rootCmd.AddCommand(failCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Panic(err)
	}
}
