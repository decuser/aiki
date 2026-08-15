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

func TestBareImportJustifiesOnlyNamedPackageRoots(t *testing.T) {
	root := minimalTree(t)
	writeTestFile(t, root, "test/behavior/import_smoke.ai", "import(\"target\", :x)\n")
	writeTestFile(t, root, "test/behavior/import_smoke.gold", "EXIT:0\n")

	local := "test/behavior/target.ai"
	writeTestFile(t, root, local, "package \"target\"\nexport(:x)\nlet x = 1\n")

	lib := "lib/target.ai"
	writeTestFile(t, root, lib, "package \"target\"\nexport(:x)\nlet x = 1\n")
	writeTestFile(t, root, "lib/target.help", "")
	writeTestFile(t, root, "lib/target.doc", "")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Orphans, local) {
		t.Fatalf("bare import must not justify arbitrary local package: %#v", result.Orphans)
	}
	if hasFinding(result.Orphans, lib) || hasFinding(result.Errors, lib) {
		t.Fatalf("bare import should justify named package root: errors=%#v orphans=%#v", result.Errors, result.Orphans)
	}
}

func TestRelativeImportedHelperIsJustified(t *testing.T) {
	root := minimalTree(t)
	writeTestFile(t, root, "test/behavior/import_smoke.ai", "import(\"./import_target\", :x)\n")
	writeTestFile(t, root, "test/behavior/import_smoke.gold", "EXIT:0\n")
	path := "test/behavior/import_target.ai"
	writeTestFile(t, root, path, "package \"import_target\"\nexport(:x)\nlet x = 1\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(result.Orphans, path) {
		t.Fatalf("relative imported helper reported as orphan: %#v", result.Orphans)
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

func TestSelfHostSourcesAndAuthorityProjectionAreJustified(t *testing.T) {
	root := minimalTree(t)
	writeTestFile(t, root, "selfhost/lexer.ai", "let lex = (source) { source }\n")
	writeTestFile(t, root, "selfhost/token_authority.ai", "println(\"authority\")\n")
	writeTestFile(t, root, "selfhost/token_authority.gold", "authority\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"selfhost/lexer.ai", "selfhost/token_authority.ai", "selfhost/token_authority.gold"} {
		if hasFinding(result.Errors, path) || hasFinding(result.Orphans, path) {
			t.Fatalf("self-host artifact should be justified: %s errors=%#v orphans=%#v", path, result.Errors, result.Orphans)
		}
	}
}

func TestSyntaxConformancePairIsJustified(t *testing.T) {
	root := minimalTree(t)
	input := "test/conformance/syntax/lex/basic.input"
	projection := "test/conformance/syntax/lex/basic.tokens"
	writeTestFile(t, root, input, "let x = 1\n")
	writeTestFile(t, root, projection, "NAME let\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{input, projection} {
		if hasFinding(result.Errors, path) || hasFinding(result.Orphans, path) {
			t.Fatalf("syntax conformance pair should be justified: %s errors=%#v orphans=%#v", path, result.Errors, result.Orphans)
		}
	}
}

func TestSyntaxConformanceInputWithoutProjectionIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	path := "test/conformance/syntax/lex/orphan.input"
	writeTestFile(t, root, path, "let x = 1\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, path) {
		t.Fatalf("expected structural error for %s, got %#v", path, result.Errors)
	}
}

func TestSyntaxConformanceProjectionWithoutInputIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	path := "test/conformance/syntax/lex/orphan.tokens"
	writeTestFile(t, root, path, "NAME let\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, path) {
		t.Fatalf("expected structural error for %s, got %#v", path, result.Errors)
	}
}

func TestXedEditorArtifactsAreJustified(t *testing.T) {
	root := minimalTree(t)
	paths := []string{
		"extra/editors/xed/aiki.lang",
		"extra/editors/xed/aiki_lsp.plugin",
		"extra/editors/xed/aiki_lsp/__init__.py",
		"extra/editors/xed/README.md",
	}
	writeTestFile(t, root, paths[0], "<language/>\n")
	writeTestFile(t, root, paths[1], "[Plugin]\nLoader=python3\nModule=aiki_lsp\n")
	writeTestFile(t, root, paths[2], "# plugin\n")
	writeTestFile(t, root, paths[3], "# Xed\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if hasFinding(result.Errors, path) || hasFinding(result.Orphans, path) {
			t.Fatalf("Xed editor artifact should be justified: %s errors=%#v orphans=%#v", path, result.Errors, result.Orphans)
		}
	}
}

func TestXedPluginDescriptorWithoutModuleIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	path := "extra/editors/xed/aiki_lsp.plugin"
	writeTestFile(t, root, path, "[Plugin]\nLoader=python3\nModule=aiki_lsp\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, path) {
		t.Fatalf("expected structural error for %s, got %#v", path, result.Errors)
	}
}

func TestXedPluginModuleWithoutDescriptorIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	path := "extra/editors/xed/aiki_lsp/__init__.py"
	writeTestFile(t, root, path, "# plugin\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, path) {
		t.Fatalf("expected structural error for %s, got %#v", path, result.Errors)
	}
}

func TestVSCodeEditorArtifactsAreJustified(t *testing.T) {
	root := minimalTree(t)
	paths := []string{
		"extra/editors/vscode/package.json",
		"extra/editors/vscode/extension.js",
		"extra/editors/vscode/language-configuration.json",
		"extra/editors/vscode/syntaxes/aiki.tmLanguage.json",
		"extra/editors/vscode/README.md",
	}
	writeTestFile(t, root, paths[0], "{}\n")
	writeTestFile(t, root, paths[1], "module.exports = {};\n")
	writeTestFile(t, root, paths[2], "{}\n")
	writeTestFile(t, root, paths[3], "{}\n")
	writeTestFile(t, root, paths[4], "# VS Code\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if hasFinding(result.Errors, path) || hasFinding(result.Orphans, path) {
			t.Fatalf("VS Code editor artifact should be justified: %s errors=%#v orphans=%#v", path, result.Errors, result.Orphans)
		}
	}
}

func TestVSCodeManifestMissingClientIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	writeTestFile(t, root, "extra/editors/vscode/package.json", "{}\n")
	writeTestFile(t, root, "extra/editors/vscode/language-configuration.json", "{}\n")
	writeTestFile(t, root, "extra/editors/vscode/syntaxes/aiki.tmLanguage.json", "{}\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, "extra/editors/vscode/package.json") {
		t.Fatalf("expected VS Code structural error, got %#v", result.Errors)
	}
}

func TestVSCodeCompanionWithoutManifestIsStructuralError(t *testing.T) {
	root := minimalTree(t)
	path := "extra/editors/vscode/extension.js"
	writeTestFile(t, root, path, "module.exports = {};\n")

	result, err := Check(root, "treecheck.allow")
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(result.Errors, path) {
		t.Fatalf("expected structural error for %s, got %#v", path, result.Errors)
	}
}
