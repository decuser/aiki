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

func TestSelfhostNewlineConformance(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)

	checkSource, err := os.ReadFile(filepath.Join(root, "selfhost", "check_normalization_authority.ai"))
	if err != nil {
		t.Fatalf("read normalization authority check: %v", err)
	}
	stdout, stderr, exitCode, err := runAikiSource(exe, string(checkSource), root)
	if err != nil {
		t.Fatalf("run normalization authority check: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("normalization authority check exited %d\nstderr: %s", exitCode, stderr)
	}
	if strings.TrimSpace(stdout) != "ok" {
		t.Fatalf("normalization authority coupling failed\nstdout: %s\nstderr: %s", stdout, stderr)
	}

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	fixtureDir := filepath.Join(root, "test", "conformance", "syntax", "newline")
	fixtures, err := filepath.Glob(filepath.Join(fixtureDir, "*.input"))
	if err != nil {
		t.Fatalf("glob newline fixtures: %v", err)
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("no newline conformance fixtures")
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
			raw, err := goLexer.Tokenize()
			if err != nil {
				t.Fatalf("Go lexer rejected fixture: %v", err)
			}
			normalized := syntax.NormalizeTokens(g, raw)
			var goLines []string
			for _, tok := range normalized {
				goLines = append(goLines, fmt.Sprintf("%d:%d %s %q", tok.Pos.Line, tok.Pos.Col, tok.Type, tok.Lexeme))
			}
			if got := strings.Join(goLines, "\n"); got != want {
				t.Fatalf("reviewed newline projection disagrees with Go normalization\nwant:\n%s\n\ngot:\n%s", want, got)
			}

			rel, err := filepath.Rel(root, fixture)
			if err != nil {
				t.Fatalf("relative fixture path: %v", err)
			}
			rel = filepath.ToSlash(rel)
			harness := "let file = import(\"file\")\n" +
				"let lexer = import(\"./selfhost/lexer\")\n" +
				"let normalizer = import(\"./selfhost/normalize\")\n" +
				"let f = file.open(" + strconv.Quote(rel) + ", :read)\n" +
				"let source = file.read_text(f)\n" +
				"file.close(f)\n" +
				"let raw = lexer.tokenize(source)\n" +
				"if is_error(raw) { println(inspect(raw)) } else { println(lexer.project_tokens(normalizer.normalize(raw))) }\n"

			stdout, stderr, exitCode, err := runAikiSource(exe, harness, root)
			if err != nil {
				t.Fatalf("run Aiki normalization: %v", err)
			}
			if exitCode != 0 {
				t.Fatalf("Aiki normalization exited %d\nstderr: %s", exitCode, stderr)
			}
			if got := strings.TrimSpace(stdout); got != want {
				t.Fatalf("reviewed newline projection disagrees with Aiki normalization\nwant:\n%s\n\ngot:\n%s\n\nstderr:\n%s", want, got, stderr)
			}
		})
	}
}
