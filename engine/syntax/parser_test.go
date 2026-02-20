// engine/syntax/parser_test.go
package syntax

import (
	"aiki/engine"
	"aiki/engine/internal"
	"regexp"
	"testing"
)

type mockParseContract struct{ internal.SilentObserver }

func (m *mockParseContract) GetTokens() []TokenDef {
	return []TokenDef{
		{Name: "WORD", Pattern: regexp.MustCompile(`^[a-z]+`)},
		{Name: "SPACE", Pattern: regexp.MustCompile(`^[ ]+`), Skip: true},
	}
}

func (m *mockParseContract) GetProduction(name string) (Production, bool) {
	if name == "root" {
		return Production{
			Expressions: [][]Term{
				{{Value: "WORD", IsSymbol: false, IsRepeat: true}},
			},
		}, true
	}
	return Production{}, false
}

func (m *mockParseContract) GetStart() string            { return "root" }
func (m *mockParseContract) Observe() engine.Observer    { return m }
func (m *mockParseContract) SetObserver(engine.Observer) {}

func TestParser(t *testing.T) {
	contract := &mockParseContract{}
	lexer := NewLexer("test.ai", "hello world", contract)
	parser, err := NewParser(lexer, contract)
	if err != nil {
		t.Fatalf("parser init failed: %v", err)
	}

	tree, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(tree.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(tree.Children))
	}
}
