package parser

import (
	"context"
	"strings"

	"github.com/RaniduNethma/vedoc/internal/models"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func ParseExpressCode(sourceCode []byte, filename string) []models.Endpoint {
	var endpoints []models.Endpoint
	var lang *sitter.Language

	if strings.HasSuffix(filename, ".ts") {
		lang = typescript.GetLanguage()
	} else {
		lang = javascript.GetLanguage()
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, _ := parser.ParseCtx(context.Background(), nil, sourceCode)
	rootNode := tree.RootNode()

	queryStr := `
	(call_expression
		(member_expression
			(property_identifier) @method
		)
		(arguments
			(string) @path
		)
	) @full_route
	`

	q, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return endpoints
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(q, rootNode)

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		var endpoint models.Endpoint

		for _, c := range m.Captures {
			nodeName := q.CaptureNameForId(c.Index)
			nodeContent := c.Node.Content(sourceCode)

			if nodeName == "method" {
				endpoint.Method = strings.ToUpper(nodeContent)
			} else if nodeName == "path" {
				endpoint.Path = strings.Trim(nodeContent, `'"`)
			} else if nodeName == "full_route" {
				endpoint.CodeSnippet = nodeContent
			}
		}

		validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true}

		if validMethods[endpoint.Method] {
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}
