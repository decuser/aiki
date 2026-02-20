// engine/syntax/lexer_test.go
package syntax

import (
	"aiki/engine"
	"aiki/engine/internal"
	"regexp"
	"testing"
)

type mockLexContract struct{ internal.SilentObserver }

func (m *mockLexContract) GetTokens() []TokenDef {
	return []TokenDef{
		{Name: "WORD", Pattern: regexp.MustCompile(`^[a-z]+`)},
		{Name: "SPACE", Pattern: regexp.MustCompile(`^[ ]+`), Skip: true},
	}
}

func (m *mockLexContract) GetProduction(string) (Production, bool) { return Production{}, false }
func (m *mockLexContract) GetStart() string                        { return "" }
func (m *mockLexContract) Observe() engine.Observer                { return m }
func (m *mockLexContract) SetObserver(engine.Observer)             {}

func TestLexer(t *testing.T) {
	contract := &mockLexContract{}
	lexer := NewLexer("test.ai", "hello world", contract)

	tok, err := lexer.NextToken()
	if err != nil || tok.Lexeme != "hello" {
		t.Errorf("lexer failed on first token: %v", err)
	}
}
