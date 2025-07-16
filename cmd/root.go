package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type rootOpts struct {
	logLevel string
}

var rootopts = rootOpts{}

var rootCmd = &cobra.Command{
	Use:   "bazeldnf",
	Short: "bazeldnf is a tool which can query RPM repos and determine package dependencies",
	Long:  `The tool allows resolving package dependencies mainly for the purpose to create custom-built SCRATCH containers consisting of RPMs, trimmed down to the absolute necessary`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func setLogLevel(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Println("Unable to parse log level from environment variable BAZELDNF_LOG_LEVEL")
		os.Exit(1)
	}
	logrus.SetLevel(level)
}

func expandParamFileArgs(args []string) []string {
	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		// bazeldnf commands can be invoked using the bazel "use_param_file"
		// approach. The parameters file, prefixed with an "@" character, is
		// expected to be the sole argument provided.
		paramFile := strings.TrimPrefix(args[0], "@")
		return readArgsFromFile(paramFile)
	}
	return args
}

func readArgsFromFile(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read params file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var contents []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		contents = append(contents, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading params file: %v\n", err)
		os.Exit(1)
	}
	return contents
}

func Execute() {
	args := expandParamFileArgs(os.Args[1:])
	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().StringVarP(&rootopts.logLevel, "log-level", "l", "", "log level")

	cobra.OnInitialize(initRootCmd)

	rootCmd.AddCommand(NewXATTRCmd())
	rootCmd.AddCommand(NewSandboxCmd())
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewLockFileCmd())
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

func initRootCmd() {
	if logLevel, hasIt := os.LookupEnv("BAZELDNF_LOG_LEVEL"); hasIt {
		setLogLevel(logLevel)
	}

	if rootopts.logLevel != "" {
		setLogLevel(rootopts.logLevel)
	}

	if chrootPath, hasIt := os.LookupEnv("BUILD_WORKING_DIRECTORY"); hasIt {
		os.Chdir(chrootPath)
	}

}
