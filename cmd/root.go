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
	rootCmd.AddCommand(NewXATTRCmd())
	rootCmd.AddCommand(NewSandboxCmd())
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewRpmTreeCmd())
	rootCmd.AddCommand(NewResolveCmd())
	rootCmd.AddCommand(NewReduceCmd())
	rootCmd.AddCommand(NewRpm2TarCmd())
	rootCmd.AddCommand(NewPruneCmd())
	rootCmd.AddCommand(NewTar2FilesCmd())
	rootCmd.AddCommand(NewLddCmd())
	rootCmd.AddCommand(NewVerifyCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
