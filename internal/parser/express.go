package parser

import (
	"context"
	"strings"

	"github.com/RaniduNethma/vedoc/internal/models"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

func ParseExpressCode(sourceCode []byte) []models.Endpoint {
	var endpoints []models.Endpoint

	lang := javascript.GetLanguage()
	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, _ := parser.ParseCtx(context.Background(), nil, sourceCode)
	rootNode := tree.RootNode()

	queryStr := `
	(call_expression
		function: (member_expression
			property: (property_identifier) @method (#match? @method "^(get|post|put|delete|patch)$")
		)
		arguments: (arguments (string) @path)
	) @full_route
	`

	q, _ := sitter.NewQuery([]byte(queryStr), lang)
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
		
		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}
