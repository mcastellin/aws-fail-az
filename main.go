package main

import (
	"log"

	_cmd "github.com/mcastellin/aws-fail-az/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aws-fail-az",
	Short: "aws-fail-az is an AWS utility to simulate Availability Zone failure",
}

var failCmd = &cobra.Command{
	Use:   "fail",
	Short: "Start AZ failure injection based on the provided configuration from stdin",
	Run: func(cmd *cobra.Command, args []string) {
		_cmd.FailCommand()
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover from AZ failure and restore saved resources state",
	Run: func(cmd *cobra.Command, args []string) {
		_cmd.RecoverCommand()
	},
}

func main() {

	rootCmd.AddCommand(failCmd)
	rootCmd.AddCommand(recoverCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Panic(err)
	}
}
