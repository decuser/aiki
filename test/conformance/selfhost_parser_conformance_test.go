//go:build rigorous

package conformance

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestSelfhostParserConformance runs the independent Aiki front end over the
// existing engine structure corpus and requires its grammar-tree projection to
// match the already reviewed *.parse.gold surface. The ordinary engine gates
// independently require the Go parser to match those same artifacts.
func TestSelfhostParserConformance(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)

	fixtureDir := filepath.Join(root, "test", "structure", "engine")
	fixtures, err := filepath.Glob(filepath.Join(fixtureDir, "*.ai"))
	if err != nil {
		t.Fatalf("glob parser fixtures: %v", err)
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("no parser conformance fixtures")
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(strings.TrimSuffix(filepath.Base(fixture), ".ai"), func(t *testing.T) {
			goldPath := fixture + ".parse.gold"
			goldBytes, err := os.ReadFile(goldPath)
			if err != nil {
				t.Fatalf("read parse projection: %v", err)
			}
			want := strings.TrimSpace(string(goldBytes))

			rel, err := filepath.Rel(root, fixture)
			if err != nil {
				t.Fatalf("relative fixture path: %v", err)
			}
			rel = filepath.ToSlash(rel)
			harness := "let file = import(\"file\")\n" +
				"let lexer = import(\"./selfhost/lexer\")\n" +
				"let normalizer = import(\"./selfhost/normalize\")\n" +
				"let parser = import(\"./selfhost/parser\")\n" +
				"let f = file.open(" + strconv.Quote(rel) + ", :read)\n" +
				"let source = file.read_text(f)\n" +
				"file.close(f)\n" +
				"let raw = lexer.tokenize(source)\n" +
				"if is_error(raw) { println(inspect(raw)) } else {\n" +
				"  let normalized = normalizer.normalize(raw)\n" +
				"  let tree = parser.parse(normalized)\n" +
				"  if is_error(tree) { println(inspect(tree)) } else { println(parser.project_tree(tree)) }\n" +
				"}\n"

			stdout, stderr, exitCode, err := runAikiSource(exe, harness, root)
			if err != nil {
				t.Fatalf("run Aiki parser: %v", err)
			}
			if exitCode != 0 {
				t.Fatalf("Aiki parser exited %d\nstderr: %s", exitCode, stderr)
			}
			if got := strings.TrimSpace(stdout); got != want {
				t.Fatalf("self-host parser projection disagrees with reviewed engine parse surface\nwant:\n%s\n\ngot:\n%s\n\nstderr:\n%s", want, got, stderr)
			}
		})
	}
}
