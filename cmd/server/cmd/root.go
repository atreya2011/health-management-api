package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Used for flags
	cfgFile     string
	configPath  string
	verboseMode bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "health-api",
	Short: "Health Management API Server",
	Long: `A backend system for a health management application following 
Domain-Driven Design (DDD) principles, built with Go, PostgreSQL, and Connect-RPC.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./configs/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", "./configs", "path to config directory")
	rootCmd.PersistentFlags().BoolVarP(&verboseMode, "verbose", "v", false, "enable verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Config is loaded by each command as needed
}
