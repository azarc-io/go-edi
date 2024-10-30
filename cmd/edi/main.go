package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set during build time
var Version = "dev" // default version if not set

var rootCmd = &cobra.Command{
	Use:     "edi",
	Short:   "A CLI tool for EDI marshaling and unmarshalling",
	Version: Version,
}

// Execute runs the root command
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig) // Initialize viper settings
	// Add the marshal and unmarshal subcommands
	rootCmd.AddCommand(marshalCmd)
	rootCmd.AddCommand(unmarshalCmd)
}

func initConfig() {
	viper.AutomaticEnv() // Automatically read environment variables
}
