/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the Bulut server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return login(args)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func login(args []string) error {
	// Check if the user is already logged in
	serverURL := getServerURL()
	apiKey, err := getApiKeyForServer(serverURL)
	if err != nil && !errors.Is(err, ring.ErrNotFound) {
		return err
	}
	if apiKey != "" {
		fmt.Println("You are already logged in!")
		fmt.Println("If you want to login with different credentials, logout first.")
		fmt.Println("To see your login info, run `bulut whoami`")
		return nil
	}

	// Get the API key from the user
	fmt.Print("Enter your API Key: ")
	_, err = fmt.Scanln(&apiKey)
	if err != nil {
		return err
	}
	err = saveToken(serverURL, apiKey)
	if err != nil {
		return err
	}
	fmt.Println("Key saved to keyring successfully!")
	return nil
}

func saveToken(serverURL, apiKey string) error {
	keyName := fmt.Sprintf("api-key_%s", serverURL)
	return ring.Set(keyName, apiKey)
}
