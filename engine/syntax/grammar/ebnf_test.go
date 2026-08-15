package grammar

import (
	"strings"
	"testing"
)

const newlineTestTokens = `@tokens {
    NEWLINE /\n+/
    NAME /[a-z]+/
    KEYWORD true false
    DELIMITER ( ) [ ]
}
`

func TestParseNewlineDeclaration(t *testing.T) {
	source := newlineTestTokens + `
@newline {
    token NEWLINE
    after_token NAME
    after_lexeme ) ] true false
    suppress_in ( )
    suppress_in [ ]
    @help "newline policy"
}
program = { NAME }
`
	g, err := NewParser("test.ebnfx", source).Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if g.Newline == nil {
		t.Fatal("Newline is nil")
	}
	if g.Newline.Token != "NEWLINE" {
		t.Fatalf("Token = %q, want NEWLINE", g.Newline.Token)
	}
	if got := strings.Join(g.Newline.AfterToken, ","); got != "NAME" {
		t.Fatalf("AfterToken = %q, want NAME", got)
	}
	if got := strings.Join(g.Newline.AfterLexeme, ","); got != "),],true,false" {
		t.Fatalf("AfterLexeme = %q", got)
	}
	if len(g.Newline.SuppressIn) != 2 || g.Newline.SuppressIn[0] != [2]string{"(", ")"} || g.Newline.SuppressIn[1] != [2]string{"[", "]"} {
		t.Fatalf("SuppressIn = %#v", g.Newline.SuppressIn)
	}
	if g.Newline.Meta.Help != "newline policy" {
		t.Fatalf("help = %q", g.Newline.Meta.Help)
	}
}

func TestParseRequiresNewlineDeclaration(t *testing.T) {
	source := newlineTestTokens + `program = { NAME }
`
	_, err := NewParser("test.ebnfx", source).Parse()
	if err == nil || !strings.Contains(err.Error(), "missing @newline declaration") {
		t.Fatalf("err = %v, want missing @newline declaration", err)
	}
}

func TestValidateNewlineDeclarationRejectsUnknownNames(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"token", "token MISSING\nafter_token NAME", `token "MISSING" is not declared`},
		{"after token", "token NEWLINE\nafter_token MISSING", `after_token "MISSING" is not declared`},
		{"after lexeme", "token NEWLINE\nafter_token NAME\nafter_lexeme ?", `after_lexeme "?" is not declared`},
		{"suppress lexeme", "token NEWLINE\nafter_token NAME\nsuppress_in { }", `suppress_in "{" "}" must name literal lexemes`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := newlineTestTokens + "\n@newline {\n" + tt.body + "\n}\nprogram = { NAME }\n"
			_, err := NewParser("test.ebnfx", source).Parse()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestValidateNewlineDeclarationRejectsDuplicatesAndConflicts(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"duplicate token directive", "token NEWLINE\ntoken NEWLINE\nafter_token NAME", "duplicate @newline token directive"},
		{"duplicate after token", "token NEWLINE\nafter_token NAME NAME", `duplicate @newline after_token "NAME"`},
		{"duplicate after lexeme", "token NEWLINE\nafter_token NAME\nafter_lexeme true true", `duplicate @newline after_lexeme "true"`},
		{"duplicate pair", "token NEWLINE\nafter_token NAME\nsuppress_in ( )\nsuppress_in ( )", `duplicate @newline suppress_in "(" ")"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := newlineTestTokens + "\n@newline {\n" + tt.body + "\n}\nprogram = { NAME }\n"
			_, err := NewParser("test.ebnfx", source).Parse()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want substring %q", err, tt.want)
			}
		})
	}
}
