package generator

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/RaniduNethma/vedoc/internal/models"
)

type OpenAPI struct {
	OpenAPI string                            `json:"openapi"`
	Info    OpenAPIInfo                       `json:"info"`
	Paths   map[string]map[string]Operation   `json:"paths"`
}

type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type RequestBody struct {
	Content map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema  Schema      `json:"schema"`
	Example interface{} `json:"example,omitempty"`
}

type Schema struct {
	Type string `json:"type"`
}

type Response struct {
	Description string `json:"description"`
}

func GenerateSwagger(endpoints []models.Endpoint, filename string) error {
	doc := OpenAPI{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:   "Vedoc Generated API (Swagger)",
			Version: "1.0.0",
		},
		Paths: make(map[string]map[string]Operation),
	}

	for _, ep := range endpoints {
		path := ep.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		method := strings.ToLower(ep.Method)
		
		tag := getFolderName(ep.Path)

		op := Operation{
			Summary:     ep.Method + " " + path,
			Description: ep.Description,
			Tags:        []string{tag},
			Responses: map[string]Response{
				"200": {Description: "Successful response"},
			},
		}

		if ep.Payload != "" && ep.Payload != "{}" && ep.Payload != `""` {
			var exampleJSON interface{}
			err := json.Unmarshal([]byte(ep.Payload), &exampleJSON)
			if err != nil {
				exampleJSON = ep.Payload
			}

			op.RequestBody = &RequestBody{
				Content: map[string]MediaType{
					"application/json": {
						Schema:  Schema{Type: "object"},
						Example: exampleJSON,
					},
				},
			}
		}

		if doc.Paths[path] == nil {
			doc.Paths[path] = make(map[string]Operation)
		}
		
		doc.Paths[path][method] = op
	}

	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}
