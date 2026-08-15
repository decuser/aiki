package treecheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, root, path, contents string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func minimalTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, "treecheck.allow", "# test allow file\n")
	return root
}

func hasFinding(findings []Finding, path string) bool {
	for _, f := range findings {
		if f.Path == path {
			return true
		}
	}
	return false
}

func TestAuditFindingsLedgerIsRepositoryArtifact(t *testing.T) {
	root := minimalTree(t)
	path := "docs/audit-findings.md"
	writeTestFile(t, root, path, "# Audit findings\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(result.Errors, path) || hasFinding(result.Orphans, path) {
		t.Fatalf("audit findings ledger should be justified: errors=%#v orphans=%#v", result.Errors, result.Orphans)
	}
}

func TestEngineGoldWithoutSpecimenIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	path := "test/structure/engine/orphan_engine.ai.lex.gold"
	writeTestFile(t, root, path, "gold\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, path) {
		t.Fatalf("expected structural error for %s, got %#v", path, result.Errors)
	}
	if hasFinding(result.Orphans, path) {
		t.Fatalf("structural error should not also be reported as orphan")
	}
}

func TestUnpairedBehaviorProgramIsOrphan(t *testing.T) {
	root := minimalTree(t)
	path := "test/behavior/old.ai"
	writeTestFile(t, root, path, "println(1)\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Orphans, path) {
		t.Fatalf("expected orphan for %s, got %#v", path, result.Orphans)
	}
}

func TestSmokePairIsJustified(t *testing.T) {
	root := minimalTree(t)
	ai := "test/behavior/example_smoke.ai"
	gold := "test/behavior/example_smoke.gold"
	writeTestFile(t, root, ai, "println(1)\n")
	writeTestFile(t, root, gold, "OUT:1\\n\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(result.Errors, ai) || hasFinding(result.Errors, gold) || hasFinding(result.Orphans, ai) || hasFinding(result.Orphans, gold) {
		t.Fatalf("paired smoke should be justified: errors=%#v orphans=%#v", result.Errors, result.Orphans)
	}
}

func TestAllowPrefixSuppressesIntentionalStandaloneTree(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "treecheck.allow", "proposals/\n")
	path := "proposals/idea.md"
	writeTestFile(t, root, path, "# idea\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(result.Orphans, path) {
		t.Fatalf("allowed path reported as orphan: %#v", result.Orphans)
	}
}

func TestImportedPackageHelperIsJustified(t *testing.T) {
	root := minimalTree(t)
	writeTestFile(t, root, "test/behavior/import_smoke.ai", "import(\"target\", :x)\n")
	writeTestFile(t, root, "test/behavior/import_smoke.gold", "EXIT:0\n")
	path := "test/behavior/target.ai"
	writeTestFile(t, root, path, "package \"target\"\nexport(:x)\nlet x = 1\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(result.Orphans, path) {
		t.Fatalf("imported helper package reported as orphan: %#v", result.Orphans)
	}
}

func TestDeletedTrackedPathIsNotTreatedAsPresent(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	writeTestFile(t, root, "treecheck.allow", "# test allow file\n")
	path := "test/behavior/removed.ai"
	writeTestFile(t, root, path, "println(1)\n")

	for _, args := range [][]string{{"init"}, {"add", "treecheck.allow", path}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(path))); err != nil {
		t.Fatal(err)
	}

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(result.Orphans, path) || hasFinding(result.Errors, path) {
		t.Fatalf("deleted tracked path should not be checked: errors=%#v orphans=%#v", result.Errors, result.Orphans)
	}
}
