package fuzz

import (
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

var testGrammar *grammar.Grammar

func init() {
	var err error
	testGrammar, err = grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		panic(err)
	}
}

// FuzzLexer tests that the lexer doesn't panic on arbitrary input.
func FuzzLexer(f *testing.F) {
	// Seed corpus with valid and edge-case inputs
	f.Add("")
	f.Add("42")
	f.Add("let x = 1")
	f.Add("\"hello\"")
	f.Add("'a'")
	f.Add(":symbol")
	f.Add("@shape")
	f.Add("[1, 2, 3]")
	f.Add("(x) { x + 1 }")
	f.Add("# comment\n42")
	f.Add("\"unterminated")
	f.Add("'")
	f.Add("@")
	f.Add(":")
	f.Add("[[[[[")
	f.Add("}}}}}}")
	f.Add("\x00\x01\x02")
	f.Add("let 你好 = 1")

	f.Fuzz(func(t *testing.T, input string) {
		// Should not panic
		lexer := syntax.NewLexer(testGrammar, "<fuzz>", input, nil)
		_, _ = lexer.Tokenize() // errors are fine, panics are not
	})
}

// FuzzParser tests that the parser doesn't panic on arbitrary input.
func FuzzParser(f *testing.F) {
	// Seed corpus
	f.Add("")
	f.Add("42")
	f.Add("let x = 1")
	f.Add("if true { 1 }")
	f.Add("while x { y }")
	f.Add("match x { 1 { a } _ { b } }")
	f.Add("(a, b) { a + b }")
	f.Add("[1, 2, 3]")
	f.Add("[@ok, 1]")
	f.Add("f(x)")
	f.Add("x.y.z")
	f.Add("a |> b |> c")
	f.Add("let let let")
	f.Add("{ { { } } }")
	f.Add("((((")

	f.Fuzz(func(t *testing.T, input string) {
		lexer := syntax.NewLexer(testGrammar, "<fuzz>", input, nil)
		tokens, err := lexer.Tokenize()
		if err != nil {
			return // lexer error is fine
		}

		// Should not panic
		parser := syntax.NewParser(testGrammar, tokens, input, nil)
		_, _ = parser.Parse() // errors are fine, panics are not
	})
}

// FuzzNumberParsing tests number parsing doesn't panic.
func FuzzNumberParsing(f *testing.F) {
	f.Add("0")
	f.Add("42")
	f.Add("3.14")
	f.Add("1/3")
	f.Add("999999999999999999999999999999")
	f.Add("1/999999999999999999999999999999")
	f.Add("0.000000000000000000000000001")
	f.Add("-42")
	f.Add("1e10")
	f.Add("1.5e-3")

	f.Fuzz(func(t *testing.T, input string) {
		lexer := syntax.NewLexer(testGrammar, "<fuzz>", input, nil)
		_, _ = lexer.Tokenize()
	})
}
