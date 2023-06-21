/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"

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
	// Ask user for confirmation of exposing the API key
	fmt.Println("\nAre you sure you want to expose your API key?")
	fmt.Println("This is not recommended if you are in a public place.")
	fmt.Println("If you are sure, type 'yes' and press enter.")
	fmt.Print("(yes/no): ")
	var confirmation string
	_, err = fmt.Scanln(&confirmation)
	if err != nil {
		return err
	}
	if confirmation != "yes" {
		fmt.Println("Aborting...")
		return nil
	}
	fmt.Println("API Key:", apiKey)
	return nil
}
