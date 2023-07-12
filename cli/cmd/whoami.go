/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"

	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the login info saved by the login command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return whoami(args)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func whoami(args []string) error {
	serverURL := getServerURL()
	apiKey, err := getApiKeyForServer(serverURL)
	if err != nil {
		if errors.Is(err, ring.ErrNotFound) {
			fmt.Println("You are not logged in!")
			return nil
		}
		return err
	}
	if apiKey == "" {
		fmt.Println("You are not logged in! (empty API key)")
		return nil
	}
	var confirmation bool
	prompt := &survey.Confirm{
		Message: "Are you sure you want to expose your API key?",
	}
	err = survey.AskOne(prompt, &confirmation)
	if err != nil {
		return err
	}
	if confirmation != true {
		fmt.Println("Aborting...")
		return nil
	}
	fmt.Println("API Key:", apiKey)
	return nil
}
