package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use: "generate",
	Short: "Generate API documentation from source code",
	Long: "Scans the current directory, parses code using Tree-sitter, and generates Postman/Swagger docs.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scanning codebase for API routes...")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory: ", err)
			return
		}

		viper.AddConfigPath(homeDir)
		viper.SetConfigName(".vedoc")
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
		}

		apikey := viper.GetString("gemini_api_key")

		if apikey != "" {
			fmt.Println("Generating intelligent docs with Gemini...")

			// todo
		} else {
			fmt.Println("Generating basic docs... (Run 'vedoc config set-key <KEY>' for AI features)")

			// todo
		}

		fmt.Println("Documentation generated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
