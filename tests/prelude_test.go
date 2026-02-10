package tests

import (
	"os"
	"testing"

	"aiki/eval"
	"aiki/value"
)

func TestPreludeRawListAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isSymbol bool
		isBool   bool
	}{
		{
			name:     "range basic",
			input:    `len(range(1, 4))`,
			expected: "3",
		},
		{
			name: "range with each",
			input: `let result = []
let collect = (n) { result = append(result, n) }
each(range(1, 4), collect)
len(result)`,
			expected: "3",
		},
		{
			name: "map over range",
			input: `let double = (n) { return n * 2 }
let doubled = map(range(1, 4), double)
nth(doubled, 0)`,
			expected: "2",
		},
		{
			name: "filter on range",
			input: `let isEven = (n) { return n % 2 == 0 }
let evens = filter(range(1, 6), isEven)
len(evens)`,
			expected: "2",
		},
		{
			name: "reduce on range",
			input: `let add = (acc, n) { return acc + n }
reduce(range(1, 6), 0, add)`,
			expected: "15",
		},
		{
			name: "find in range",
			input: `let gt5 = (n) { return n > 5 }
		find(range(1, 10), gt5)`,
			expected: "6",
		},
		{
			name: "any on range",
			input: `let gt3 = (n) { return n > 3 }
any(range(1, 5), gt3)`,
			expected: "true",
			isBool:   true,
		},
		{
			name: "all on range",
			input: `let lt10 = (n) { return n < 10 }
all(range(1, 5), lt10)`,
			expected: "true",
			isBool:   true,
		},
		{
			name: "reverse raw list",
			input: `let rev = reverse(range(1, 4))
nth(rev, 0)`,
			expected: "3",
		},
		{
			name:     "sum of range",
			input:    `sum(range(1, 6))`,
			expected: "15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := value.NewEnv(nil)

			// Load prelude
			preludeSource := loadPreludeSourceForTest()
			result := eval.Run(preludeSource, env)
			if e, ok := result.(*value.Error); ok {
				t.Fatalf("prelude load error: %s", e.Message)
			}

			result = eval.Run(tt.input, env)

			if tt.isSymbol {
				sym, ok := result.(*value.Symbol)
				if !ok {
					if e, ok := result.(*value.Error); ok {
						t.Fatalf("expected Symbol, got Error: %s", e.Message)
					}
					t.Fatalf("expected Symbol, got %T: %v", result, result)
				}
				if sym.Value != tt.expected {
					t.Errorf("got :%s, want :%s", sym.Value, tt.expected)
				}
			} else if tt.isBool {
				b, ok := result.(*value.Boolean)
				if !ok {
					if e, ok := result.(*value.Error); ok {
						t.Fatalf("expected Boolean, got Error: %s", e.Message)
					}
					t.Fatalf("expected Boolean, got %T: %v", result, result)
				}
				expectedBool := tt.expected == "true"
				if b.Value != expectedBool {
					t.Errorf("got %v, want %v", b.Value, expectedBool)
				}
			} else {
				num, ok := result.(*value.Number)
				if !ok {
					if e, ok := result.(*value.Error); ok {
						t.Fatalf("expected Number, got Error: %s", e.Message)
					}
					t.Fatalf("expected Number, got %T: %v", result, result)
				}
				if num.Inspect() != tt.expected {
					t.Errorf("got %s, want %s", num.Inspect(), tt.expected)
				}
			}
		})
	}
}

func loadPreludeSourceForTest() string {
	data, err := os.ReadFile("../prelude/prelude.ai")
	if err != nil {
		panic("failed to load prelude: " + err.Error())
	}
	return string(data)
}
