package invariant

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestSelfhostLexerConformance(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)

	// First prove that the independent lexer's duplicated enumerated tables are
	// still exactly the lexical facts derived by Aiki from grammar.ebnfx.
	checkSource, err := os.ReadFile(filepath.Join(root, "selfhost", "check_lexer_authority.ai"))
	if err != nil {
		t.Fatalf("read lexer authority check: %v", err)
	}
	stdout, stderr, exitCode, err := runAikiSource(exe, string(checkSource), root)
	if err != nil {
		t.Fatalf("run lexer authority check: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("lexer authority check exited %d\nstderr: %s", exitCode, stderr)
	}
	if strings.TrimSpace(stdout) != "ok" {
		t.Fatalf("lexer authority coupling failed\nstdout: %s\nstderr: %s", stdout, stderr)
	}

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	fixtureDir := filepath.Join(root, "test", "conformance", "syntax", "lex")
	fixtures, err := filepath.Glob(filepath.Join(fixtureDir, "*.input"))
	if err != nil {
		t.Fatalf("glob lexical fixtures: %v", err)
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("no lexical conformance fixtures")
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(strings.TrimSuffix(filepath.Base(fixture), ".input"), func(t *testing.T) {
			sourceBytes, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			goldPath := strings.TrimSuffix(fixture, ".input") + ".tokens"
			goldBytes, err := os.ReadFile(goldPath)
			if err != nil {
				t.Fatalf("read fixture projection: %v", err)
			}
			want := strings.TrimSpace(string(goldBytes))

			goLexer := syntax.NewLexer(g, filepath.Base(fixture), string(sourceBytes), nil)
			goTokens, err := goLexer.Tokenize()
			if err != nil {
				t.Fatalf("Go lexer rejected fixture: %v", err)
			}
			var goLines []string
			for _, tok := range goTokens {
				goLines = append(goLines, fmt.Sprintf("%d:%d %s %q", tok.Pos.Line, tok.Pos.Col, tok.Type, tok.Lexeme))
			}
			goProjection := strings.Join(goLines, "\n")
			if goProjection != want {
				t.Fatalf("reviewed token projection disagrees with Go lexer\nwant:\n%s\n\ngot:\n%s", want, goProjection)
			}

			rel, err := filepath.Rel(root, fixture)
			if err != nil {
				t.Fatalf("relative fixture path: %v", err)
			}
			rel = filepath.ToSlash(rel)
			harness := "let file = import(\"file\")\n" +
				"let lexer = import(\"./selfhost/lexer\")\n" +
				"let f = file.open(" + strconv.Quote(rel) + ", :read)\n" +
				"let source = file.read_text(f)\n" +
				"file.close(f)\n" +
				"let tokens = lexer.tokenize(source)\n" +
				"if is_error(tokens) { println(inspect(tokens)) } else { println(lexer.project_tokens(tokens)) }\n"

			stdout, stderr, exitCode, err := runAikiSource(exe, harness, root)
			if err != nil {
				t.Fatalf("run Aiki lexer: %v", err)
			}
			if exitCode != 0 {
				t.Fatalf("Aiki lexer exited %d\nstderr: %s", exitCode, stderr)
			}
			aikiProjection := strings.TrimSpace(stdout)
			if aikiProjection != want {
				t.Fatalf("reviewed token projection disagrees with Aiki lexer\nwant:\n%s\n\ngot:\n%s\n\nstderr:\n%s", want, aikiProjection, stderr)
			}
		})
	}
}
