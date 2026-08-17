package conformance

import (
	"strconv"
	"strings"
	"testing"
)

// TestSelfhostEvaluatorExpressionConformance requires the Aiki-written
// evaluator to agree with the reference evaluator on expression semantics.
// The two implementations share language values through the ordinary prelude,
// but do not share evaluator implementation code.
func TestSelfhostEvaluatorExpressionConformance(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)

	cases := []string{
		"1 + 2 * 3", // Aiki is deliberately left-to-right: (1+2)*3 == 9.
		"-3 + 5",
		"not false",
		"true and 7",
		"false or 7",
		"\"a\" + \"b\"",
		"\"a\\n\"",
		"'A'",
		"'\\n'",
		":dynamic",
		"[1, 2, 3][1]",
		"(1 + 2) * 3",
	}

	for _, expr := range cases {
		expr := expr
		t.Run(expr, func(t *testing.T) {
			direct := "println(inspect(" + expr + "))\n"
			wantOut, wantErr, wantCode, err := runAikiSource(exe, direct, root)
			if err != nil {
				t.Fatalf("run reference evaluator: %v", err)
			}
			if wantCode != 0 {
				t.Fatalf("reference evaluator exited %d\nstderr: %s", wantCode, wantErr)
			}

			source := expr + "\n"
			harness := "let lexer = import(\"./selfhost/lexer\")\n" +
				"let normalizer = import(\"./selfhost/normalize\")\n" +
				"let parser = import(\"./selfhost/parser\")\n" +
				"let evaluator = import(\"./selfhost/evaluator\")\n" +
				"let runtime = import(\"./selfhost/runtime\")\n" +
				"let @node [kind, value, children, line, col]\n" +
				"let tree = parser.parse(normalizer.normalize(lexer.tokenize(" + strconv.Quote(source) + ")))\n" +
				"let expr = tree.children[0].children[0].children[0]\n" +
				"println(inspect(evaluator.eval_expr(expr, runtime.env_new([]))))\n"

			gotOut, gotErr, gotCode, err := runAikiSource(exe, harness, root)
			if err != nil {
				t.Fatalf("run self-host evaluator: %v", err)
			}
			if gotCode != 0 {
				t.Fatalf("self-host evaluator exited %d\nstderr: %s", gotCode, gotErr)
			}
			if strings.TrimSpace(gotOut) != strings.TrimSpace(wantOut) {
				t.Fatalf("expression result differs\nreference: %q\nself-host: %q\nself-host stderr: %s", strings.TrimSpace(wantOut), strings.TrimSpace(gotOut), gotErr)
			}
		})
	}
}
