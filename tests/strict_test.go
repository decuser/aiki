package tests

import (
	"strings"
	"testing"

	"aiki/strict"
)

func TestRange(t *testing.T) {
	result := testEvalStrict(`range(1, 5)`)
	if result.Inspect() != "[1, 2, 3, 4]" {
		t.Errorf("got %s, want [1, 2, 3, 4]", result.Inspect())
	}
}

func TestSum(t *testing.T) {
	result := testEvalStrict(`sum([1, 2, 3, 4, 5])`)
	testNumberValue(t, result, "15")
}

func TestMap(t *testing.T) {
	result := testEvalStrict(`map([1, 2, 3], (x) { return x * 2 })`)
	if result.Inspect() != "[2, 4, 6]" {
		t.Errorf("got %s, want [2, 4, 6]", result.Inspect())
	}
}

func TestFilter(t *testing.T) {
	result := testEvalStrict(`filter([1, 2, 3, 4, 5], (x) { return x > 3 })`)
	if result.Inspect() != "[4, 5]" {
		t.Errorf("got %s, want [4, 5]", result.Inspect())
	}
}

func TestReduce(t *testing.T) {
	result := testEvalStrict(`reduce([1, 2, 3, 4], 0, (acc, x) { return acc + x })`)
	testNumberValue(t, result, "10")
}

func TestReverse(t *testing.T) {
	result := testEvalStrict(`reverse([1, 2, 3])`)
	if result.Inspect() != "[3, 2, 1]" {
		t.Errorf("got %s, want [3, 2, 1]", result.Inspect())
	}
}

func TestFind(t *testing.T) {
	result := testEvalStrict(`find([10, 20, 30], (x) { return x > 15 })`)
	testNumberValue(t, result, "20")
}

func TestAnyAll(t *testing.T) {
	result := testEvalStrict(`any([1, 2, 3], (x) { return x > 2 })`)
	testBooleanValue(t, result, true)

	result = testEvalStrict(`all([1, 2, 3], (x) { return x > 0 })`)
	testBooleanValue(t, result, true)

	result = testEvalStrict(`all([1, 2, 3], (x) { return x > 2 })`)
	testBooleanValue(t, result, false)
}

func TestMinMax(t *testing.T) {
	result := testEvalStrict(`min([3, 1, 4, 1, 5])`)
	testNumberValue(t, result, "1")

	result = testEvalStrict(`max([3, 1, 4, 1, 5])`)
	testNumberValue(t, result, "5")
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
			result := testEvalStrict(tt.input)
			if result.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", result.Inspect(), tt.expected)
			}
		})
	}
}

// TestStrictExportsMatch verifies that strict.Exports() matches
// the export [...] line in strict.ai source.
func TestStrictExportsMatch(t *testing.T) {
	source := strict.Source
	goExports := strict.Exports()

	sourceExports := extractExportNames(source)
	if len(sourceExports) == 0 {
		t.Fatal("no export statement found in strict.ai")
	}

	goSet := make(map[string]bool)
	for _, name := range goExports {
		goSet[name] = true
	}

	sourceSet := make(map[string]bool)
	for _, name := range sourceExports {
		sourceSet[name] = true
	}

	for _, name := range goExports {
		if !sourceSet[name] {
			t.Errorf("Exports() has '%s' but strict.ai export does not", name)
		}
	}

	for _, name := range sourceExports {
		if !goSet[name] {
			t.Errorf("strict.ai exports '%s' but Exports() does not", name)
		}
	}

	if len(goExports) != len(sourceExports) {
		t.Errorf("count mismatch: Exports() has %d, strict.ai has %d",
			len(goExports), len(sourceExports))
	}
}

// TestStrictExportsAreDefined verifies every exported name has a let binding.
func TestStrictExportsAreDefined(t *testing.T) {
	source := strict.Source
	exports := strict.Exports()

	for _, name := range exports {
		pattern := "let " + name + " ="
		if !strings.Contains(source, pattern) {
			// Also check for "let name = " with space before =
			pattern2 := "let " + name + " ="
			if !strings.Contains(source, pattern2) {
				t.Errorf("exported name '%s' has no 'let %s = ...' in strict.ai", name, name)
			}
		}
	}
}

func extractExportNames(source string) []string {
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "export [") {
			continue
		}
		start := strings.Index(line, "[")
		end := strings.Index(line, "]")
		if start < 0 || end < 0 || end <= start {
			continue
		}
		inner := line[start+1 : end]
		parts := strings.Split(inner, ",")
		var names []string
		for _, p := range parts {
			name := strings.TrimSpace(p)
			if name != "" {
				names = append(names, name)
			}
		}
		return names
	}
	return nil
}
