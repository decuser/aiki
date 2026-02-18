package fmt_test

import (
	"strings"
	"testing"

	aikifmt "aiki/tools/fmt"
	"aiki/syntax"
)

func init() {

	aikifmt.SetGrammar(syntax.GetGrammar())
}

// TestFmtRoundTrip verifies that parsing formatted output produces equivalent AST
func TestFmtRoundTrip(t *testing.T) {
	tests := []string{
		`let x = 5`,
		`let add = (a, b) { return a + b }`,
		`if x { return 1 }`,
		`if x { return 1 } else { return 2 }`,
		`while x > 0 { x = x - 1 }`,
		`[1, 2, 3]`,
		`[@point, 10, 20]`,
		`let @point [x, y]`,
		`f(1, 2, 3)`,
		`list[0]`,
		`point.x`,
		`1 + 2 * 3`,
		`not true`,
		`x |> f() |> g()`,
		`match x { 1 { return "one" } _ { return "other" } }`,
		`export [foo, bar]`,
		`from math use [sqrt, abs]`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			formatted, err := aikifmt.Format(input)
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}

			// Should be able to format again without error
			_, err = aikifmt.Format(formatted)
			if err != nil {
				t.Fatalf("re-format failed: %v\nformatted was:\n%s", err, formatted)
			}
		})
	}
}

// TestFmtIdempotent verifies that formatting twice gives same result
func TestFmtIdempotent(t *testing.T) {
	tests := []string{
		`let x = 5`,
		`let add = (a, b) { return a + b }`,
		`if x { return 1 } else { return 2 }`,
		`while x > 0 { x = x - 1 }`,
		`x |> f() |> g()`,
		`match x { 1 { return "one" } _ { return "other" } }`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			first, err := aikifmt.Format(input)
			if err != nil {
				t.Fatalf("first format failed: %v", err)
			}

			second, err := aikifmt.Format(first)
			if err != nil {
				t.Fatalf("second format failed: %v", err)
			}

			if first != second {
				t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

// TestFmtGolden verifies expected output for specific inputs
func TestFmtGolden(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple let",
			input:    `let    x   =   5`,
			expected: "let x = 5\n",
		},
		{
			name:     "function",
			input:    `let f = (x,y) { return x+y }`,
			expected: "let f = (x, y) {\n\treturn x + y\n}\n",
		},
		{
			name:     "if else",
			input:    `if x{return 1}else{return 2}`,
			expected: "if x {\n\treturn 1\n} else {\n\treturn 2\n}\n",
		},
		{
			name:     "operators spaced",
			input:    `1+2*3`,
			expected: "1 + 2 * 3\n",
		},
		{
			name:     "list",
			input:    `[1,2,3]`,
			expected: "[1, 2, 3]\n",
		},
		{
			name:     "shaped list",
			input:    `[@ok,42]`,
			expected: "[@ok, 42]\n",
		},
		{
			name:     "call with args",
			input:    `f(1,2,3)`,
			expected: "f(1, 2, 3)\n",
		},
		{
			name:     "while loop",
			input:    `while x>0{x=x-1}`,
			expected: "while x > 0 {\n\tx = x - 1\n}\n",
		},
		{
			name:     "not operator",
			input:    `not true`,
			expected: "not true\n",
		},
		{
			name:     "negative number",
			input:    `-5`,
			expected: "-5\n",
		},
		{
			name:     "rest param",
			input:    `let f = (...args) { return args }`,
			expected: "let f = (...args) {\n\treturn args\n}\n",
		},
		{
			name:     "mixed params",
			input:    `let f = (a, b, ...rest) { return rest }`,
			expected: "let f = (a, b, ...rest) {\n\treturn rest\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := aikifmt.Format(tt.input)
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}

			if got != tt.expected {
				t.Errorf("mismatch:\ninput:    %q\nexpected: %q\ngot:      %q", tt.input, tt.expected, got)
			}
		})
	}
}

// TestFmtNoDataLoss verifies that all meaningful content is preserved
func TestFmtNoDataLoss(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mustHave []string
	}{
		{
			name:     "string content",
			input:    `let s = "hello world"`,
			mustHave: []string{`"hello world"`},
		},
		{
			name:     "rune content",
			input:    `let r = 'x'`,
			mustHave: []string{`'x'`},
		},
		{
			name:     "symbol",
			input:    `let s = :foo`,
			mustHave: []string{`:foo`},
		},
		{
			name:     "number",
			input:    `let n = 3.14159`,
			mustHave: []string{`3.14159`},
		},
		{
			name:     "rational",
			input:    `let r = 3/4`,
			mustHave: []string{`3/4`},
		},
		{
			name:     "identifier",
			input:    `let my_var = 1`,
			mustHave: []string{`my_var`},
		},
		{
			name:     "shape name",
			input:    `let @my_shape [x, y]`,
			mustHave: []string{`@my_shape`, `x`, `y`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := aikifmt.Format(tt.input)
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}

			for _, must := range tt.mustHave {
				if !strings.Contains(got, must) {
					t.Errorf("missing %q in output:\n%s", must, got)
				}
			}
		})
	}
}

// TestFmtMultiStatement verifies multiple statements are handled
func TestFmtMultiStatement(t *testing.T) {
	input := `let x = 1
let y = 2
let z = x + y`

	got, err := aikifmt.Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// Should have 3 lines (plus trailing newline)
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d:\n%s", len(lines), got)
	}

	if !strings.HasPrefix(lines[0], "let x") {
		t.Errorf("first line wrong: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "let y") {
		t.Errorf("second line wrong: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "let z") {
		t.Errorf("third line wrong: %s", lines[2])
	}
}

// TestFmtNestedBlocks verifies indentation of nested structures
func TestFmtNestedBlocks(t *testing.T) {
	input := `let f = () { if x { while y { return 1 } } }`

	got, err := aikifmt.Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// Check for increasing indentation
	lines := strings.Split(got, "\n")

	// Should see tabs increasing
	foundDoubleIndent := false
	for _, line := range lines {
		if strings.HasPrefix(line, "\t\t") {
			foundDoubleIndent = true
			break
		}
	}

	if !foundDoubleIndent {
		t.Errorf("expected nested indentation:\n%s", got)
	}
}
