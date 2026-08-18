//go:build rigorous

package conformance

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestSelfhostPreludeBridgeCoverage prevents the explicit host-value bridge
// from becoming a second authority for the prelude vocabulary.
func TestSelfhostPreludeBridgeCoverage(t *testing.T) {
	root := distributionRoot(t)
	preludeBytes, err := os.ReadFile(filepath.Join(root, "engine", "runtime", "prelude", "prelude.ai"))
	if err != nil {
		t.Fatal(err)
	}
	bridgeBytes, err := os.ReadFile(filepath.Join(root, "selfhost", "host_prelude.ai"))
	if err != nil {
		t.Fatal(err)
	}

	decl := regexp.MustCompile(`(?m)^let ([A-Za-z_][A-Za-z0-9_]*) = \(`)
	var want []string
	for _, m := range decl.FindAllStringSubmatch(string(preludeBytes), -1) {
		want = append(want, m[1])
	}

	bridge := string(bridgeBytes)
	start := strings.Index(bridge, "let names = [")
	end := strings.Index(bridge[start:], "]")
	if start < 0 || end < 0 {
		t.Fatal("cannot locate host prelude names inventory")
	}
	block := bridge[start : start+end+1]
	quoted := regexp.MustCompile(`"([A-Za-z_][A-Za-z0-9_]*)"`)
	var got []string
	for _, m := range quoted.FindAllStringSubmatch(block, -1) {
		got = append(got, m[1])
	}

	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("self-host prelude bridge drift\nwant: %v\ngot:  %v", want, got)
	}
}
