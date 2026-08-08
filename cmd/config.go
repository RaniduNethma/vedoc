package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use: "config",
	Short: "Manage Vedoc Configurations",
	Long: "Set or update your Gemini API key and other configurations globally.",
}

var setKeyCmd = &cobra.Command{
	Use: "set-key [API_KEY]",
	Short: "Set your Gemini api key globally",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var apiKey string

		// If user provides the key as an argument, warn them but accept it
		if len(args) == 1 {
			fmt.Println("Warning: Passing the API key as a command-line argument may expose it in your shell history.")
			apiKey = args[0]
		} else {
			fmt.Print("Enter Gemini API Key: ")
			byteKey, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Println("\nError reading API key:", err)
				return
			}
			fmt.Println()
			apiKey = strings.TrimSpace(string(byteKey))
		}

		if apiKey == "" {
			fmt.Println("Error: API key cannot be empty.")
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return
		}

		configPath := filepath.Join(homeDir, ".vedoc.yml")

		viper.Set("gemini_api_key", apiKey)

		err = viper.WriteConfigAs(configPath)
		if err != nil {
			fmt.Println("Error saving API key: ", err)
			return
		}

		fmt.Printf("Gemini API Key saved successfully to %s\n", configPath)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setKeyCmd)
}
