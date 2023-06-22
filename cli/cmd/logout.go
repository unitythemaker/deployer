/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the token saved by the login command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return logout(args)
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func logout(args []string) error {
	serverURL := getServerURL()
	err := ring.Delete(fmt.Sprintf("api-key_%s", serverURL))
	if err != nil {
		if errors.Is(err, ring.ErrNotFound) {
			fmt.Println("You are not logged in!")
			return nil
		}
		return err
	}
	fmt.Println("Logged out successfully!")
	return nil
}
