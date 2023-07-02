/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

// namespaceCmd represents the namespace command
var namespaceCmd = &cobra.Command{
	Use:     "namespace",
	Aliases: []string{"ns"},
	Short:   "Manage namespaces",
	Long:    `Namespace is a way to divide cluster resources between multiple projects.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return checkLogin()
	},
}

var nsCreateCmd = &cobra.Command{
	Use:   "create [namespace]",
	Short: "Create a namespace",
	RunE: func(cmd *cobra.Command, args []string) error {
		return createNamespaceHandler(args)
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
	namespaceCmd.AddCommand(nsCreateCmd)
}

func createNamespaceHandler(args []string) error {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		prompt := &survey.Input{
			Message: "New namespace name",
		}
		err := survey.AskOne(prompt, &name)
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("namespace name cannot be empty")
		}
	}
	fmt.Printf("Creating namespace \"%s\"", name)
	return nil
}

func createNamespace(name string) error {
	serverURL := getServerURL()
	apiKey, err := getApiKeyForServer(serverURL)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", serverURL+"/api/v1/namespace", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errResp := &bytes.Buffer{}
		_, err := io.Copy(errResp, resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("bad status: %s, response: %s", resp.Status, errResp)
	}

	return nil
}

// TODO: Move to a common place
func checkLogin() error {
	serverURL := getServerURL()
	_, err := getApiKeyForServer(serverURL)
	if err != nil {
		return err
	}
	return nil
}
