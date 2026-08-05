package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RaniduNethma/vedoc/internal/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type BatchGeminiResponse struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Payload     string `json:"payload"`
}

func EnrichEndpointsBatch(apikey string, endpoints []models.Endpoint) ([]models.Endpoint, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apikey))
	if err != nil {
		return endpoints, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-3.5-flash")

	var endpointsInfo strings.Builder
	for _, ep := range endpoints {
		endpointsInfo.WriteString(fmt.Sprintf("Method: %s\nPath: %s\nCode:\n%s\n---\n", ep.Method, ep.Path, ep.CodeSnippet))
	}

	prompt := fmt.Sprintf(`Act as an expert API technical writer. 
    Analyze the following list of API endpoints and their code snippets. Provide a brief description and a sample JSON request payload for each.

    CRITICAL RULES:
    1. Respond ONLY with a valid JSON array of objects.
    2. The objects must have EXACTLY the keys: "method", "path", "description", and "payload".
    3. The "payload" value MUST be a stringified JSON (escaped as a string). If no payload is needed, use an empty string "".
    4. Do NOT include markdown tags like %s or backticks.

    Endpoints:
    %s`, "```json", endpointsInfo.String())

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return endpoints, err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		textResp := fmt.Sprintf("%v", part)

		textResp = strings.ReplaceAll(textResp, "```json", "")
		textResp = strings.ReplaceAll(textResp, "```", "")
		textResp = strings.TrimSpace(textResp)

		var aiResults []BatchGeminiResponse
		err := json.Unmarshal([]byte(textResp), &aiResults)
		if err != nil {
			fmt.Println("AI JSON Parse Error:", err)
			return endpoints, err
		}

		for i, originalEp := range endpoints {
			for _, aiEp := range aiResults {
				if strings.EqualFold(originalEp.Method, aiEp.Method) && originalEp.Path == aiEp.Path {
					endpoints[i].Description = aiEp.Description
					endpoints[i].Payload = aiEp.Payload
					break
				}
			}
		}
	}

	return endpoints, nil
}
