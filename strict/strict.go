package strict

import (
	_ "embed"
	"strings"
)

//go:embed strict.ai
var Source string

// Exports returns the exported names from strict.ai by parsing
// the export statement directly from source. This is the single
// source of truth — no hardcoded list to fall out of sync.
func Exports() []string {
	lines := strings.Split(Source, "\n")
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
