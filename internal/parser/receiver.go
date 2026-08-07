package parser

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type expressSymbolKind uint8

const (
	expressSymbolUnknown expressSymbolKind = iota
	expressFactory
	expressRouterFactory
	expressApp
	expressRouter
)

type expressBinding struct {
	position uint32
	kind     expressSymbolKind
}

type expressScope struct {
	parent   *expressScope
	bindings map[string][]expressBinding
}

type expressReceiverAnalyzer struct {
	source       []byte
	root         *expressScope
	scopesByNode map[uintptr]*expressScope
}

var (
	defaultExpressImportPattern   = regexp.MustCompile(`(?s)^\s*import\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:,|from)`) // default import
	namespaceExpressImportPattern = regexp.MustCompile(`(?s)\*\s+as\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	namedExpressImportPattern     = regexp.MustCompile(`(?s)\{([^}]*)\}`)
)

func newExpressReceiverAnalyzer(root *sitter.Node, source []byte) *expressReceiverAnalyzer {
	rootScope := newExpressScope(nil)
	analyzer := &expressReceiverAnalyzer{
		source:       source,
		root:         rootScope,
		scopesByNode: map[uintptr]*expressScope{root.ID(): rootScope},
	}
	analyzer.walk(root, rootScope)
	return analyzer
}

func newExpressScope(parent *expressScope) *expressScope {
	return &expressScope{parent: parent, bindings: make(map[string][]expressBinding)}
}

func (a *expressReceiverAnalyzer) provesReceiver(node *sitter.Node) bool {
	if node == nil || node.IsNull() {
		return false
	}
	scope := a.scopeForNode(node)
	kind := a.classifyExpression(node, scope, node.StartByte())
	return kind == expressApp || kind == expressRouter
}

func (a *expressReceiverAnalyzer) scopeForNode(node *sitter.Node) *expressScope {
	for current := node; current != nil && !current.IsNull(); current = current.Parent() {
		if scope, ok := a.scopesByNode[current.ID()]; ok {
			return scope
		}
	}
	return a.root
}

func (a *expressReceiverAnalyzer) walk(node *sitter.Node, scope *expressScope) {
	if node == nil || node.IsNull() {
		return
	}

	if node.Type() != "program" && createsExpressScope(node.Type()) {
		childScope := newExpressScope(scope)
		a.scopesByNode[node.ID()] = childScope
		if isFunctionScope(node.Type()) {
			a.bindFunctionParameters(node, childScope)
		} else if node.Type() == "catch_clause" {
			a.bindCatchParameter(node, childScope)
		}
		for i := 0; i < int(node.NamedChildCount()); i++ {
			a.walk(node.NamedChild(i), childScope)
		}
		return
	}

	switch node.Type() {
	case "import_statement":
		a.bindExpressImport(node, scope)
		return
	case "variable_declarator":
		value := nonNullNode(node.ChildByFieldName("value"))
		if value != nil {
			a.walk(value, scope)
		}
		a.bindVariableDeclarator(node, value, scope)
		return
	case "assignment_expression":
		right := nonNullNode(node.ChildByFieldName("right"))
		if right != nil {
			a.walk(right, scope)
		}
		a.bindAssignment(node, right, scope)
		return
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		a.walk(node.NamedChild(i), scope)
	}
}

func createsExpressScope(nodeType string) bool {
	if isFunctionScope(nodeType) {
		return true
	}
	switch nodeType {
	case "statement_block", "if_statement", "switch_statement", "for_statement", "for_in_statement", "while_statement", "do_statement", "try_statement", "catch_clause":
		return true
	default:
		return false
	}
}

func isFunctionScope(nodeType string) bool {
	switch nodeType {
	case "function_declaration", "function_expression", "arrow_function", "generator_function_declaration", "generator_function", "method_definition":
		return true
	default:
		return false
	}
}

func (a *expressReceiverAnalyzer) bindFunctionParameters(node *sitter.Node, scope *expressScope) {
	parameters := nonNullNode(node.ChildByFieldName("parameters"))
	if parameters == nil {
		parameter := nonNullNode(node.ChildByFieldName("parameter"))
		if parameter != nil {
			a.bindPatternIdentifiers(parameter, scope, expressSymbolUnknown, node.StartByte())
		}
		return
	}
	a.bindPatternIdentifiers(parameters, scope, expressSymbolUnknown, node.StartByte())
}

func (a *expressReceiverAnalyzer) bindCatchParameter(node *sitter.Node, scope *expressScope) {
	parameter := nonNullNode(node.ChildByFieldName("parameter"))
	if parameter != nil {
		a.bindPatternIdentifiers(parameter, scope, expressSymbolUnknown, node.StartByte())
	}
}

func (a *expressReceiverAnalyzer) bindPatternIdentifiers(node *sitter.Node, scope *expressScope, kind expressSymbolKind, position uint32) {
	if node == nil || node.IsNull() {
		return
	}
	if node.Type() == "identifier" {
		scope.bind(node.Content(a.source), kind, position)
		return
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		a.bindPatternIdentifiers(node.NamedChild(i), scope, kind, position)
	}
}

func (a *expressReceiverAnalyzer) bindExpressImport(node *sitter.Node, scope *expressScope) {
	sourceNode := nonNullNode(node.ChildByFieldName("source"))
	if sourceNode == nil || trimStringLiteral(sourceNode.Content(a.source)) != "express" {
		return
	}

	content := node.Content(a.source)
	position := node.EndByte()

	if match := defaultExpressImportPattern.FindStringSubmatch(content); len(match) == 2 && match[1] != "type" {
		scope.bind(match[1], expressFactory, position)
	}
	if match := namespaceExpressImportPattern.FindStringSubmatch(content); len(match) == 2 {
		scope.bind(match[1], expressFactory, position)
	}
	if match := namedExpressImportPattern.FindStringSubmatch(content); len(match) == 2 {
		for _, part := range strings.Split(match[1], ",") {
			fields := strings.Fields(strings.TrimSpace(part))
			if len(fields) == 0 || fields[0] != "Router" {
				continue
			}
			name := "Router"
			if len(fields) >= 3 && fields[1] == "as" {
				name = fields[2]
			}
			scope.bind(name, expressRouterFactory, position)
		}
	}
}

func (a *expressReceiverAnalyzer) bindVariableDeclarator(node, value *sitter.Node, scope *expressScope) {
	name := nonNullNode(node.ChildByFieldName("name"))
	if name == nil {
		return
	}

	position := node.EndByte()
	if name.Type() == "identifier" {
		kind := expressSymbolUnknown
		if value != nil {
			kind = a.classifyExpression(value, scope, value.StartByte())
		}
		scope.bind(name.Content(a.source), kind, position)
		return
	}

	if name.Type() == "object_pattern" && value != nil && a.isRequireExpress(value) {
		a.bindRouterDestructure(name, scope, position)
	}
}

func (a *expressReceiverAnalyzer) bindRouterDestructure(pattern *sitter.Node, scope *expressScope, position uint32) {
	content := strings.TrimSpace(pattern.Content(a.source))
	content = strings.TrimPrefix(content, "{")
	content = strings.TrimSuffix(content, "}")
	for _, part := range strings.Split(content, ",") {
		part = strings.TrimSpace(part)
		if part == "Router" {
			scope.bind("Router", expressRouterFactory, position)
			continue
		}
		fields := strings.Fields(part)
		if len(fields) == 3 && fields[0] == "Router" && fields[1] == ":" {
			scope.bind(fields[2], expressRouterFactory, position)
			continue
		}
		if strings.HasPrefix(part, "Router:") {
			alias := strings.TrimSpace(strings.TrimPrefix(part, "Router:"))
			if alias != "" {
				scope.bind(alias, expressRouterFactory, position)
			}
		}
	}
}

func (a *expressReceiverAnalyzer) bindAssignment(node, right *sitter.Node, scope *expressScope) {
	left := nonNullNode(node.ChildByFieldName("left"))
	if left == nil || left.Type() != "identifier" {
		return
	}

	kind := expressSymbolUnknown
	if right != nil {
		kind = a.classifyExpression(right, scope, right.StartByte())
	}
	name := left.Content(a.source)
	position := node.EndByte()

	owner := scope.findOwner(name, node.StartByte())
	if owner == nil || owner == scope {
		scope.bind(name, kind, position)
		return
	}

	// Assignments crossing a lexical/control-flow boundary are not guaranteed to
	// execute before later code in the outer scope. Invalidate the outer proof,
	// but keep the branch-local value available for calls inside this scope.
	owner.bind(name, expressSymbolUnknown, position)
	scope.bind(name, kind, position)
}

func (a *expressReceiverAnalyzer) classifyExpression(node *sitter.Node, scope *expressScope, position uint32) expressSymbolKind {
	if node == nil || node.IsNull() {
		return expressSymbolUnknown
	}

	switch node.Type() {
	case "identifier":
		return scope.lookup(node.Content(a.source), position)
	case "call_expression":
		if a.isRequireExpress(node) {
			return expressFactory
		}
		function := nonNullNode(node.ChildByFieldName("function"))
		if function == nil {
			return expressSymbolUnknown
		}
		switch a.classifyExpression(function, scope, function.StartByte()) {
		case expressFactory:
			return expressApp
		case expressRouterFactory:
			return expressRouter
		default:
			return expressSymbolUnknown
		}
	case "member_expression":
		object := nonNullNode(node.ChildByFieldName("object"))
		property := nonNullNode(node.ChildByFieldName("property"))
		if object == nil || property == nil || property.Content(a.source) != "Router" {
			return expressSymbolUnknown
		}
		if a.classifyExpression(object, scope, object.StartByte()) == expressFactory {
			return expressRouterFactory
		}
		return expressSymbolUnknown
	case "parenthesized_expression", "as_expression", "type_assertion", "non_null_expression", "satisfies_expression":
		if node.NamedChildCount() > 0 {
			child := node.NamedChild(0)
			return a.classifyExpression(child, scope, child.StartByte())
		}
	}

	return expressSymbolUnknown
}

func (a *expressReceiverAnalyzer) isRequireExpress(node *sitter.Node) bool {
	if node == nil || node.IsNull() || node.Type() != "call_expression" {
		return false
	}
	function := nonNullNode(node.ChildByFieldName("function"))
	arguments := nonNullNode(node.ChildByFieldName("arguments"))
	if function == nil || arguments == nil || function.Type() != "identifier" || function.Content(a.source) != "require" {
		return false
	}
	if arguments.NamedChildCount() != 1 {
		return false
	}
	argument := arguments.NamedChild(0)
	return argument.Type() == "string" && trimStringLiteral(argument.Content(a.source)) == "express"
}

func (s *expressScope) bind(name string, kind expressSymbolKind, position uint32) {
	if name == "" {
		return
	}
	s.bindings[name] = append(s.bindings[name], expressBinding{position: position, kind: kind})
}

func (s *expressScope) lookup(name string, position uint32) expressSymbolKind {
	if s == nil {
		return expressSymbolUnknown
	}
	if bindings := s.bindings[name]; len(bindings) > 0 {
		for i := len(bindings) - 1; i >= 0; i-- {
			if bindings[i].position <= position {
				return bindings[i].kind
			}
		}
	}
	return s.parent.lookup(name, position)
}

func (s *expressScope) findOwner(name string, position uint32) *expressScope {
	for current := s; current != nil; current = current.parent {
		if bindings := current.bindings[name]; len(bindings) > 0 {
			for i := len(bindings) - 1; i >= 0; i-- {
				if bindings[i].position <= position {
					return current
				}
			}
		}
	}
	return nil
}

func nonNullNode(node *sitter.Node) *sitter.Node {
	if node == nil || node.IsNull() {
		return nil
	}
	return node
}

func trimStringLiteral(value string) string {
	return strings.Trim(value, `'"`)
}
