package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSelfhostTokenAuthority proves that the Aiki-native Cut I.0 extractor can
// read the authoritative grammar and emit the reviewed deterministic lexical
// projection kept beside it. The extractor derives facts from grammar.ebnfx;
// the gold file is the independently reviewed expectation used to catch both
// extractor mistakes and grammar-surface changes.
func TestSelfhostTokenAuthority(t *testing.T) {
	root := distributionRoot(t)
	sourcePath := filepath.Join(root, "selfhost", "token_authority.ai")
	goldPath := filepath.Join(root, "selfhost", "token_authority.gold")

	gold, err := os.ReadFile(goldPath)
	if err != nil {
		t.Fatalf("read token authority gold: %v", err)
	}

	exe := buildAiki(t)
	stdout, stderr, exitCode, err := runAikiFile(exe, sourcePath, root)
	if err != nil {
		t.Fatalf("run token authority extractor: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("token authority extractor exited %d\nstderr: %s", exitCode, stderr)
	}

	got := strings.TrimSpace(stdout)
	want := strings.TrimSpace(string(gold))
	if got != want {
		t.Fatalf("token authority projection drifted\nwant:\n%s\n\ngot:\n%s\n\nstderr:\n%s", want, got, stderr)
	}
}
