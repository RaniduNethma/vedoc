package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RaniduNethma/vedoc/internal/models"
)

type postmanCollection struct {
	Info Info `json:"info"`
	Item []Item `json:"item"`
}

type Info struct {
	Name string `json:"name"`
	Schema string `json:"schema"`
}

type Item struct {
	Name string `json:"name"`
	Request Request `json:"request"`
}

type Request struct {
	Method string `json:"method"`
	Description string `json:"description,omitempty"`
	Header []Header `json:"header,omitempty"`
	Body *Body `json:"body,omitempty"`
	Url Url `json:"url"`
}

type Header struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

type Body struct {
	Mode string `json:"mode"`
	Raw string `json:"raw"`
	Options struct {
		Raw struct {
			Language string `json:"language"`
		} `json:"raw"`
	} `json:"options"`
}

type Url struct {
	Raw string `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

func GeneratePostmanCollection(endpoints []models.Endpoint, Filename string) error {
	collection := postmanCollection{
		Info: Info{
			Name: "Vedoc Generated API",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
	}

	for _, ep := range endpoints {
		pathParts := strings.Split(strings.TrimPrefix(ep.Path, "/"), "/")

		item := Item{
			Name: fmt.Sprintf("%s %s", ep.Method, ep.Path),
			Request: Request{
				Method: ep.Method,
				Description: ep.Description,
				Url: Url{
					Raw: "{{baseUrl}}" + ep.Path,
					Host: []string{"{{baseUrl}}"},
					Path: pathParts,
				},
			},
		}

		if ep.Payload != "" && ep.Payload != "{}" && ep.Payload != `""` {
			item.Request.Header = append(item.Request.Header, Header{Key: "Content-Type", Value: "application/json"})
			item.Request.Body = &Body{
				Mode: "raw",
				Raw: ep.Payload,
			}
			item.Request.Body.Options.Raw.Language = "json"
		}

		collection.Item = append(collection.Item, item)
	}

	jsonData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(Filename, jsonData, 0644)
}
