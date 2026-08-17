package conformance

import (
	"strconv"
	"strings"
	"testing"
)

// TestSelfhostEvaluatorProgramConformance exercises statement/control-flow and
// function semantics through the independent Aiki evaluator and compares the
// observable result with the reference evaluator.
func TestSelfhostEvaluatorProgramConformance(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)

	cases := []struct {
		name  string
		setup string
		probe string
	}{
		{"assign", "let x = 1\nx = x + 2\n", "x"},
		{"if true", "let x = 1\nif true { x = 7 } else { x = 9 }\n", "x"},
		{"if false", "let x = 1\nif false { x = 7 } else { x = 9 }\n", "x"},
		{"while", "let x = 0\nwhile x < 4 { x = x + 1 }\n", "x"},
		{"shape field", "let @point [x, y]\nlet p = [@point, 10, 20]\n", "p.x"},
		{"match", "let out = 0\nmatch 3 { 3 { out = 7 } _ { out = 9 } }\n", "out"},
		{"list pattern", "let out = 0\nmatch [1, 2] { [a, b] { out = a + b } _ { out = 9 } }\n", "out"},
		{"recoverable error pattern", "let e = [@error, :io, \"oops\"]\nlet out = 0\nmatch e { [@error, :io, msg] { out = length(msg) } _ { out = 9 } }\n", "out"},
		{"closures and prelude", "let add = (a, b) { return a + b }\nlet make = (x) { return (y) { return x + y } }\nlet inc = make(10)\nlet a = add(2, 3)\nlet b = inc(7)\nlet c = length([1, 2, 3])\n", "a + b + c"},
		{"recursion rest pipeline", "let fact = (n) { if n <= 1 { return 1 } return n * fact(n - 1) }\nlet sum = (first, ...rest) { let out = first let i = 0 while i < length(rest) { out = out + rest[i] i = i + 1 } return out }\nlet twice = (x) { return x * 2 }\nlet plus = (x, n) { return x + n }\nlet a = fact(5)\nlet b = sum(1, 2, 3, 4)\nlet c = 3 |> twice |> plus(5)\n", "a + b + c"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			direct := tc.setup + "println(inspect(" + tc.probe + "))\n"
			wantOut, wantErr, wantCode, err := runAikiSource(exe, direct, root)
			if err != nil {
				t.Fatalf("run reference evaluator: %v", err)
			}
			if wantCode != 0 {
				t.Fatalf("reference evaluator exited %d\nstderr: %s", wantCode, wantErr)
			}

			source := tc.setup + tc.probe + "\n"
			harness := "let lexer = import(\"./selfhost/lexer\")\n" +
				"let normalizer = import(\"./selfhost/normalize\")\n" +
				"let parser = import(\"./selfhost/parser\")\n" +
				"let evaluator = import(\"./selfhost/evaluator\")\n" +
				"let runtime = import(\"./selfhost/runtime\")\n" +
				"let host = import(\"./selfhost/host_prelude\")\n" +
				"let tree = parser.parse(normalizer.normalize(lexer.tokenize(" + strconv.Quote(source) + ")))\n" +
				"let env = runtime.env_new([])\n" +
				"host.install(env)\n" +
				"println(inspect(evaluator.eval_program(tree, env)))\n"

			gotOut, gotErr, gotCode, err := runAikiSource(exe, harness, root)
			if err != nil {
				t.Fatalf("run self-host evaluator: %v", err)
			}
			if gotCode != 0 {
				t.Fatalf("self-host evaluator exited %d\nstderr: %s", gotCode, gotErr)
			}
			if strings.TrimSpace(gotOut) != strings.TrimSpace(wantOut) {
				t.Fatalf("program result differs\nreference: %q\nself-host: %q\nself-host stderr: %s", strings.TrimSpace(wantOut), strings.TrimSpace(gotOut), gotErr)
			}
		})
	}
}
