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

type GeminiResponse struct {
	Description string `json:"description"`
	Payload string `json:"payload"`
}

func EnrichEndpoint(apikey string, ep models.Endpoint) (models.Endpoint, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apikey))
	if err != nil {
		return ep, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-3.5-flash")

	prompt := fmt.Sprintf(`Act as an expert API technical writer. 
        Analyze the following API endpoint code and provide a brief description and a sample JSON request payload.
        CRITICAL RULES:
        1. Respond ONLY with a valid JSON object.
        2. The keys must be EXACTLY "description" and "payload".
        3. The "payload" value MUST be a stringified JSON (escaped as a string), NOT a nested object. If no payload is needed, use an empty string "".
        4. Do NOT include markdown tags like %s or backticks.
        
        Example Output:
        {"description": "Creates a new user", "payload": "{\n  \"email\": \"test@test.com\"\n}"}

        Code:
        %s`, "```json", ep.CodeSnippet)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return ep, err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		textResp := fmt.Sprintf("%v", part)

		textResp = strings.ReplaceAll(textResp, "```json", "")
        textResp = strings.ReplaceAll(textResp, "```", "")
        textResp = strings.TrimSpace(textResp)

		var gResp GeminiResponse
		err := json.Unmarshal([]byte(textResp), &gResp)
		if err != nil {
			fmt.Printf("AI JSON Parse Error on [%s] %s: %v\n", ep.Method, ep.Path, err)
            fmt.Println("Raw AI Output:", textResp)
			
		} else {
			ep.Description = gResp.Description
			ep.Payload = gResp.Payload
		}
	}

	return ep, nil
}
