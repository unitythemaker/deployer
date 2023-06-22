/*
Copyright © 2023 Halil Tezcan KARABULUT <unity@themaker.cyou>
*/
package cmd

import (
	"bulut-cli/util/keyring"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var (
	cfgFile string
	appName = "BulutCLI-ProductDevBook"
	ring    keyring.Keyring
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bulut",
	Short: "bulut is a cli tool for deploying nitro apps blazing fast!",
	Long:  `Bulut is a simple deployment tool for your projects. Lightning-fast deployment.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initKeyring)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bulut.yaml)")
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "URL of the Bulut Server")

	viper.BindPFlag("server", rootCmd.Flags().Lookup("server"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Check if there is a config in the current directory
		currentDir, err := os.Getwd()
		cobra.CheckErr(err)
		for i := 0; i < 2; i++ {
			var newDir string
			var configFound bool
			configFound, newDir = checkConfigAndUpdateDir(currentDir)
			if configFound {
				viper.AddConfigPath(currentDir)
				currentDir = newDir
				break
			}
			currentDir = newDir
		}

		cobra.CheckErr(err)
	}

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".bulut")

	// Set defaults
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.ssl", false) // TODO: in the production, this should be true
	viper.SetDefault("config.build-path", ".output")
	viper.SetDefault("config.entrypoint", "server/index.mjs")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.MergeInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		fmt.Println("Using server:", getServerURL())
	}
}

func initKeyring() {
	var err error
	ring, err = keyring.Open(appName)
	cobra.CheckErr(err)
}

func checkConfigAndUpdateDir(currentDir string) (bool, string) {
	_, err := os.Stat(currentDir + "/.bulut.yaml")
	updatedDir := updateDir(currentDir)
	if err == nil {
		return true, updatedDir
	}
	return false, updatedDir
}

func updateDir(currentDir string) string {
	updatedDir, err := filepath.Abs(filepath.Join(currentDir, ".."))
	cobra.CheckErr(err)
	return updatedDir
}
