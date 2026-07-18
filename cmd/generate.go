package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RaniduNethma/vedoc/internal/ai"
	"github.com/RaniduNethma/vedoc/internal/generator"
	"github.com/RaniduNethma/vedoc/internal/models"
	"github.com/RaniduNethma/vedoc/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate API documentation from source code",
	Long:  "Scans the current directory, parses code using Tree-sitter, and generates Postman/Swagger docs.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scanning codebase for API routes...")

		var endpoints []models.Endpoint

		err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				if d.Name() == "node_modules" || (strings.HasPrefix(d.Name(), ".") && d.Name() != ".") {
					return filepath.SkipDir
				}
			}

			if !d.IsDir() && (strings.HasSuffix(d.Name(), ".js") || strings.HasSuffix(d.Name(), ".ts")) {
				sourceCode, err := os.ReadFile(path)
				if err == nil {
					fileEndpoints := parser.ParseExpressCode(sourceCode, d.Name())
					endpoints = append(endpoints, fileEndpoints...)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Println("Error scanning directory:", err)
			return
		}

		if len(endpoints) == 0 {
			fmt.Println("No API endpoints found in the current directory.")
			return
		}

		fmt.Printf("\n Found %d endpoints across the project:\n", len(endpoints))
		for _, ep := range endpoints {
			fmt.Printf("[%s] %s\n", ep.Method, ep.Path)
		}

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

			for i, ep := range endpoints {
				fmt.Printf("Analyzing [%s] %s ...\n", ep.Method, ep.Path)

				enrichedEp, err := ai.EnrichEndpoint(apikey, ep)
				if err == nil {
					endpoints[i] = enrichedEp
				} else {
					fmt.Printf("AI API Error: %v\n", err)
				}
			}
		} else {
			fmt.Println("Generating basic docs... (Run 'vedoc config set-key <KEY>' for AI features)")
		}

		fmt.Println("Documentation generated successfully!")
		fmt.Println("\n Final Documented Endpoints: ")

		for _, ep := range endpoints {
			fmt.Printf("\n [%s] %s\n", ep.Method, ep.Path)
			if ep.Description != "" {
				fmt.Printf("Description: %s\n", ep.Description)
			}
			if ep.Payload != "" && ep.Payload != "{}" && ep.Payload != `""` {
				fmt.Printf("Payload: %s\n", ep.Payload)
			}
		}

		fmt.Println("\n Generating Postman Collection...")
		err = generator.GeneratePostmanCollection(endpoints, "vedoc_postman_collection.json")

		if err != nil {
			fmt.Println("Error generating Postman collection: ", err)
		} else {
			fmt.Println("Successfully created 'vedoc_postman_collection.json'!")
			fmt.Println("You can now import this file directly into Postman.")
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
