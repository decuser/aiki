package refactor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
	refsyntax "aiki/reference/syntax"
)

// TestParserCompare compares new engine parser output against reference.
func TestParserCompare(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("finding project root: %v", err)
	}

	// Load grammar for new engine
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}

	// Get reference grammar
	refGrammar := refsyntax.GetGrammar()

	// Test files
	testFiles := []string{
		"tests/smoke/math_smoke.ai",
		"tests/smoke/functional_smoke.ai",
		"tests/smoke/iter_smoke.ai",
		"tests/smoke/pipeline_smoke.ai",
	}

	for _, file := range testFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			fullPath := filepath.Join(root, file)
			source, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("reading %s: %v", file, err)
			}

			compareParserOutput(t, g, refGrammar, file, string(source))
		})
	}
}

// compareParserOutput parses with both parsers and compares ASTs.
func compareParserOutput(t *testing.T, newGrammar *grammar.Grammar, refGrammar *refsyntax.Grammar, file, source string) {
	// Parse with new engine
	newLexer := syntax.NewLexer(newGrammar, file, source, nil)
	newTokens, err := newLexer.Tokenize()
	if err != nil {
		t.Fatalf("new lexer error: %v", err)
	}
	newParser := syntax.NewParser(newGrammar, newTokens, source, nil)
	newAST, err := newParser.Parse()
	if err != nil {
		t.Fatalf("new parser error: %v", err)
	}

	// Parse with reference
	refAST, err := refGrammar.ParseSource(source)
	if err != nil {
		t.Fatalf("ref parser error: %v", err)
	}

	// Compare ASTs
	if err := compareNodes(newAST, refAST, ""); err != nil {
		t.Errorf("AST mismatch: %v", err)
	}
}

// compareNodes recursively compares two AST nodes.
func compareNodes(newNode *syntax.Node, refNode *refsyntax.Node, path string) error {
	if newNode == nil && refNode == nil {
		return nil
	}
	if newNode == nil {
		return fmt.Errorf("%s: new is nil, ref is %s", path, refNode.Type)
	}
	if refNode == nil {
		return fmt.Errorf("%s: new is %s, ref is nil", path, newNode.Type)
	}

	// Compare type
	if newNode.Type != refNode.Type {
		return fmt.Errorf("%s: type mismatch: new=%s, ref=%s", path, newNode.Type, refNode.Type)
	}

	// Compare value (for terminals)
	if newNode.Value != refNode.Value {
		return fmt.Errorf("%s: value mismatch: new=%q, ref=%q", path, newNode.Value, refNode.Value)
	}

	// Compare children count
	if len(newNode.Children) != len(refNode.Children) {
		return fmt.Errorf("%s (%s): children count mismatch: new=%d, ref=%d",
			path, newNode.Type, len(newNode.Children), len(refNode.Children))
	}

	// Compare each child
	for i := range newNode.Children {
		childPath := fmt.Sprintf("%s/%s[%d]", path, newNode.Type, i)
		if err := compareNodes(newNode.Children[i], refNode.Children[i], childPath); err != nil {
			return err
		}
	}

	return nil
}
