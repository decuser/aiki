package tests

import (
	"strings"
	"testing"

	"aiki/strict"
)

// TestStrictExportsMatch verifies that strict.Exports() matches
// the export [...] line in strict.ai source.
func TestStrictExportsMatch(t *testing.T) {
	source := strict.Source
	goExports := strict.Exports()

	// Extract export list from source
	sourceExports := extractExportNames(source)
	if len(sourceExports) == 0 {
		t.Fatal("no export statement found in strict.ai")
	}

	// Check that Go list matches source list
	goSet := make(map[string]bool)
	for _, name := range goExports {
		goSet[name] = true
	}

	sourceSet := make(map[string]bool)
	for _, name := range sourceExports {
		sourceSet[name] = true
	}

	// Find names in Go but not in source
	for _, name := range goExports {
		if !sourceSet[name] {
			t.Errorf("strict.Exports() has '%s' but strict.ai export does not", name)
		}
	}

	// Find names in source but not in Go
	for _, name := range sourceExports {
		if !goSet[name] {
			t.Errorf("strict.ai exports '%s' but strict.Exports() does not", name)
		}
	}

	// Check counts match
	if len(goExports) != len(sourceExports) {
		t.Errorf("count mismatch: strict.Exports() has %d, strict.ai has %d",
			len(goExports), len(sourceExports))
	}
}

// TestStrictExportsAreDefined verifies that every exported name
// is actually defined as a let binding in strict.ai.
func TestStrictExportsAreDefined(t *testing.T) {
	source := strict.Source
	exports := strict.Exports()

	for _, name := range exports {
		// Look for "let name = " pattern
		pattern := "let " + name + " = "
		if !strings.Contains(source, pattern) {
			t.Errorf("exported name '%s' has no 'let %s = ...' in strict.ai", name, name)
		}
	}
}

// extractExportNames parses "export [name1, name2, ...]" from source.
func extractExportNames(source string) []string {
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "export [") {
			continue
		}
		// Extract between [ and ]
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
