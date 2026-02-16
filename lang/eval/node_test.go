package eval

import (
	"testing"

	"aiki/internal/ebnf"
	"aiki/lang/value"
	"aiki/runtime/prelude"
)

func TestEvalNodeBasic(t *testing.T) {
	g, err := ebnf.ParseFile("../../cmd/aiki/grammar.ebnf")
	if err != nil {
		t.Fatalf("parse grammar: %v", err)
	}

	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"number", "42", "42"},
		{"add", "1 + 2", "3"},
		{"sub", "5 - 3", "2"},
		{"mul", "3 * 4", "12"},
		{"div", "10 / 4", "5/2"},
		{"let", "let x = 5\nx", "5"},
		{"let-expr", "let x = 1 + 2\nx", "3"},
		{"bool-true", "true", "true"},
		{"bool-false", "false", "false"},
		{"modulo", "modulo(7, 3)", "1"},
		{"modulo-zero", "modulo(10, 5)", "0"},
		{"equal-true", "equal(5, 5)", "true"},
		{"equal-false", "equal(5, 3)", "false"},
		{"not-equal-true", "not equal(5, 3)", "true"},
		{"not-equal-false", "not equal(5, 5)", "false"},
		{"equal-string", `equal("a", "a")`, "true"},
		{"equal-list", "equal([1, 2], [1, 2])", "true"},
		{"equal-list-false", "equal([1, 2], [1, 3])", "false"},
		{"string", `"hello"`, `"hello"`},
		{"list", "[1, 2, 3]", "[1, 2, 3]"},
		{"index", "let a = [10, 20, 30]\na[1]", "20"},
		{"if-true", "if true { 1 }", "1"},
		{"if-false", "if false { 1 }", "null"},
		{"if-else", "if false { 1 } else { 2 }", "2"},
		{"compare-lt", "1 < 2", "true"},
		{"compare-gt", "2 > 1", "true"},
		{"func-call", "let f = (x) { return x + 1 }\nf(5)", "6"},
		{"func-two-args", "let add = (a, b) { return a + b }\nadd(3, 4)", "7"},
		{"while", "let x = 3\nlet sum = 0\nwhile x > 0 { sum = sum + x\nx = x - 1 }\nsum", "6"},
		{"factorial", `let factorial = (n) {
			if n <= 1 { return 1 }
			return n * factorial(n - 1)
		}
		factorial(5)`, "120"},
		{"pipe-filter", "[1,2,3,4] |> filter((x) { return x > 2 }) |> length()", "2"},
		{"pipe-map", "[1,2,3] |> map((x) { return x * 2 }) |> sum()", "12"},
		{"pipe-chain", "range(1, 6) |> filter((x) { return equal(modulo(x, 2), 0) }) |> map((x) { return x * x }) |> sum()", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := g.ParseSource(tt.source)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			env := value.NewEnv(nil)
			RunNode(g, prelude.Source, env)
			env.SnapshotStrict()

			result := EvalNode(ast, env)

			got := result.Inspect()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
