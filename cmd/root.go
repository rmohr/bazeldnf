package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bazeldnf",
	Short: "bazeldnf is a tool which can query RPM repos and determine package dependencies",
	Long:  `The tool allows resolving package dependencies mainly for the purpose to create custom-built SCRATCH containers consisting of RPMs, trimmed down to the absolute necessary`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewrpmtreeCmd())
	rootCmd.AddCommand(NewResolveCmd())
	rootCmd.AddCommand(NewReduceCmd())
	rootCmd.AddCommand(NewRPMCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
