// Package refactor contains comparison tests for the engine refactor.
// These tests verify the new engine matches the reference implementation.
// Delete this package when the refactor is complete.
package refactor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
	refsyntax "aiki/reference/syntax"
	enginesyntax "aiki/engine/syntax"
)

// findProjectRoot walks up from cwd to find go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root")
		}
		dir = parent
	}
}

// TestLexerCompare compares new engine lexer output against reference.
func TestLexerCompare(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
	    t.Fatalf("finding project root: %v", err)
	}

	// Load grammar for new engine
	g, err := grammar.Load("grammar.ebnfx", enginesyntax.EbnfxSource, "grammar.help", enginesyntax.HelpSource)
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

			compareLexerOutput(t, g, refGrammar, file, string(source))
		})
	}
}

// compareLexerOutput tokenizes with both lexers and compares.
func compareLexerOutput(t *testing.T, newGrammar *grammar.Grammar, refGrammar *refsyntax.Grammar, file, source string) {
	// Tokenize with new engine
	newLexer := syntax.NewLexer(newGrammar, file, source, nil)
	newTokens, err := newLexer.Tokenize()
	if err != nil {
		t.Fatalf("new lexer error: %v", err)
	}

	// Tokenize with reference
	refTokens, _ := refGrammar.Tokenize(source)

	// Compare token counts
	if len(newTokens) != len(refTokens) {
		t.Errorf("token count mismatch: new=%d, ref=%d", len(newTokens), len(refTokens))
		t.Logf("new tokens: %v", formatTokens(newTokens))
		t.Logf("ref tokens: %v", formatRefTokens(refTokens))
		return
	}

	// Compare each token
	for i := range newTokens {
		newTok := newTokens[i]
		refTok := refTokens[i]

		// Compare type (may need mapping)
		newType := normalizeType(newTok.Type)
		refType := normalizeRefType(refTok.Type)

		if newType != refType {
			t.Errorf("token %d type mismatch: new=%s, ref=%s (lexeme=%q)",
				i, newTok.Type, refTok.Type, newTok.Lexeme)
		}

		if newTok.Lexeme != refTok.Lexeme {
			t.Errorf("token %d lexeme mismatch: new=%q, ref=%q",
				i, newTok.Lexeme, refTok.Lexeme)
		}

		// Compare position
		if newTok.Pos.Line != refTok.Line || newTok.Pos.Col != refTok.Column {
			t.Errorf("token %d position mismatch: new=%d:%d, ref=%d:%d (lexeme=%q)",
				i, newTok.Pos.Line, newTok.Pos.Col, refTok.Line, refTok.Column, newTok.Lexeme)
		}
	}
}

// normalizeType maps new engine token types to common names.
func normalizeType(typ string) string {
	// Map to common names for comparison
	switch typ {
	case "KEYWORD":
		return "keyword"
	case "NAME":
		return "name"
	case "NUMBER":
		return "number"
	case "STRING":
		return "string"
	case "RUNE":
		return "rune"
	case "SYMBOL":
		return "symbol"
	case "SHAPE":
		return "shape"
	case "OPERATOR":
		return "operator"
	case "DELIMITER":
		return "delimiter"
	case "EOF":
		return "eof"
	default:
		return typ
	}
}

// normalizeRefType maps reference token types to common names.
func normalizeRefType(typ string) string {
	// Reference uses uppercase
	switch typ {
	case "KEYWORD":
		return "keyword"
	case "NAME":
		return "name"
	case "NUMBER":
		return "number"
	case "STRING":
		return "string"
	case "RUNE":
		return "rune"
	case "SYMBOL":
		return "symbol"
	case "SHAPE":
		return "shape"
	case "OPERATOR":
		return "operator"
	case "DELIMITER":
		return "delimiter"
	case "EOF":
		return "eof"
	default:
		return typ
	}
}

func formatTokens(tokens []syntax.Token) string {
	var result string
	for i, tok := range tokens {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s:%q", tok.Type, tok.Lexeme)
	}
	return result
}

func formatRefTokens(tokens []refsyntax.Token) string {
	var result string
	for i, tok := range tokens {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s:%q", tok.Type, tok.Lexeme)
	}
	return result
}
