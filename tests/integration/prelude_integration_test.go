package integration_test

import (
	"aiki/runtime/prelude"
	"aiki/tests/testutil"
	"strings"
	"testing"
)

func TestRange(t *testing.T) {
	result := testutil.EvalPrelude(`range(1, 5)`)
	if result.Inspect() != "[1, 2, 3, 4]" {
		t.Errorf("got %s, want [1, 2, 3, 4]", result.Inspect())
	}
}

func TestSum(t *testing.T) {
	result := testutil.EvalPrelude(`sum([1, 2, 3, 4, 5])`)
	testutil.TestNumberValue(t, result, "15")
}

func TestMap(t *testing.T) {
	result := testutil.EvalPrelude(`map([1, 2, 3], (x) { return x * 2 })`)
	if result.Inspect() != "[2, 4, 6]" {
		t.Errorf("got %s, want [2, 4, 6]", result.Inspect())
	}
}

func TestFilter(t *testing.T) {
	result := testutil.EvalPrelude(`filter([1, 2, 3, 4, 5], (x) { return x > 3 })`)
	if result.Inspect() != "[4, 5]" {
		t.Errorf("got %s, want [4, 5]", result.Inspect())
	}
}

func TestReduce(t *testing.T) {
	result := testutil.EvalPrelude(`reduce([1, 2, 3, 4], 0, (acc, x) { return acc + x })`)
	testutil.TestNumberValue(t, result, "10")
}

func TestReverse(t *testing.T) {
	result := testutil.EvalPrelude(`reverse([1, 2, 3])`)
	if result.Inspect() != "[3, 2, 1]" {
		t.Errorf("got %s, want [3, 2, 1]", result.Inspect())
	}
}

func TestFind(t *testing.T) {
	result := testutil.EvalPrelude(`find([10, 20, 30], (x) { return x > 15 })`)
	testutil.TestNumberValue(t, result, "20")
}

func TestAnyAll(t *testing.T) {
	result := testutil.EvalPrelude(`any([1, 2, 3], (x) { return x > 2 })`)
	testutil.TestBooleanValue(t, result, true)

	result = testutil.EvalPrelude(`all([1, 2, 3], (x) { return x > 0 })`)
	testutil.TestBooleanValue(t, result, true)

	result = testutil.EvalPrelude(`all([1, 2, 3], (x) { return x > 2 })`)
	testutil.TestBooleanValue(t, result, false)
}

func TestMinMax(t *testing.T) {
	result := testutil.EvalPrelude(`min([3, 1, 4, 1, 5])`)
	testutil.TestNumberValue(t, result, "1")

	result = testutil.EvalPrelude(`max([3, 1, 4, 1, 5])`)
	testutil.TestNumberValue(t, result, "5")
}

func TestHashMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"new", `type(hash_new())`, ":list"},
		{"put_get", `
let h = hash_new()
let h = hash_put(h, "key", 42)
hash_get(h, "key")
`, "42"},
		{"has", `
let h = hash_new()
let h = hash_put(h, "key", 1)
hash_has(h, "key")
`, "true"},
		{"has_missing", `
let h = hash_new()
hash_has(h, "nope")
`, "false"},
		{"del", `
let h = hash_new()
let h = hash_put(h, "key", 1)
let h = hash_del(h, "key")
hash_has(h, "key")
`, "false"},
		{"keys", `
let h = hash_new()
let h = hash_put(h, "a", 1)
let h = hash_put(h, "b", 2)
length(hash_keys(h))
`, "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutil.EvalPrelude(tt.input)
			if result.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", result.Inspect(), tt.expected)
			}
		})
	}
}

// TestPreludeExportsAreDefined verifies every exported name is defined and accessible.
func TestPreludeExportsAreDefined(t *testing.T) {
	// Evaluate prelude.ai and get exports from env
	env := testutil.SetupEnv()
	exports := env.GetExports()

	if len(exports) == 0 {
		t.Fatal("prelude.ai has no exports")
	}

	// Verify each export is actually defined
	for _, name := range exports {
		if _, ok := env.Get(name); !ok {
			t.Errorf("exported name '%s' is not defined in prelude.ai", name)
		}
	}
}

// TestPreludeExportsComplete verifies all expected functions are exported.
func TestPreludeExportsComplete(t *testing.T) {
	env := testutil.SetupEnv()
	exports := env.GetExports()

	// Expected exports
	expected := []string{
		"each", "map", "filter", "reduce", "range", "reverse",
		"find", "any", "all", "sum", "max", "min",
		"hash_new", "hash_get", "hash_put", "hash_has",
		"hash_del", "hash_keys", "hash_values", "hash_code",
		"println",
	}

	exportSet := make(map[string]bool)
	for _, name := range exports {
		exportSet[name] = true
	}

	for _, name := range expected {
		if !exportSet[name] {
			t.Errorf("expected export '%s' not found", name)
		}
	}
}

// TestPreludeSourceHasLetBindings verifies every export has a let binding in source.
func TestPreludeSourceHasLetBindings(t *testing.T) {
	source := prelude.Source
	env := testutil.SetupEnv()
	exports := env.GetExports()

	for _, name := range exports {
		pattern := "let " + name + " ="
		if !strings.Contains(source, pattern) {
			t.Errorf("exported name '%s' has no 'let %s = ...' in prelude.ai", name, name)
		}
	}
}
