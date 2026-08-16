package distfmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatGoExpandsLongPolicyLiteralAndIsStable(t *testing.T) {
	src := `package p

var policy = map[string][]string{
	"lib/file/file.ai": {"_file_close", "_file_copy", "_file_delete", "_file_exists", "_file_list", "_file_mkdir", "_file_mkdir_all", "_file_open"},
}
`
	out, err := formatGo("policy.go", src)
	if err != nil {
		t.Fatalf("formatGo: %v", err)
	}
	for _, want := range []string{"\t\"lib/file/file.ai\": {\n", "\t\t\"_file_close\",\n", "\t\t\"_file_open\",\n", "\t},\n"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	out2, err := formatGo("policy.go", out)
	if err != nil {
		t.Fatalf("second formatGo: %v", err)
	}
	if out2 != out {
		t.Fatalf("distfmt not stable\n--- first ---\n%s\n--- second ---\n%s", out, out2)
	}
}

func TestFormatGoKeepsShortPolicyLiteralCompact(t *testing.T) {
	src := `package p

var policy = map[string][]string{
	"lib/x.ai": {"_a", "_b"},
}
`
	out, err := formatGo("policy.go", src)
	if err != nil {
		t.Fatalf("formatGo: %v", err)
	}
	if !strings.Contains(out, `"lib/x.ai": {"_a", "_b"},`) {
		t.Fatalf("short literal expanded unexpectedly:\n%s", out)
	}
}

func TestFormatAikiExpandsLongListAndPreservesCanonicalStructure(t *testing.T) {
	src := "let xs = [alpha, beta, gamma, delta, epsilon, zeta, eta, theta, iota, kappa, lambda, mu, nu, xi, omicron]\n"
	out, err := formatAiki("test.ai", src)
	if err != nil {
		t.Fatalf("formatAiki: %v", err)
	}
	if !strings.Contains(out, "let xs = [\n\talpha,\n") {
		t.Fatalf("long list did not expand:\n%s", out)
	}
	out2, err := formatAiki("test.ai", out)
	if err != nil {
		t.Fatalf("second formatAiki: %v", err)
	}
	if out2 != out {
		t.Fatalf("distfmt not stable\n--- first ---\n%s\n--- second ---\n%s", out, out2)
	}
}

func TestFormatPathUsesAtomicWriteAndBackupSurface(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	src := `package p

var policy = map[string][]string{
	"lib/file/file.ai": {"_file_close", "_file_copy", "_file_delete", "_file_exists", "_file_list", "_file_mkdir", "_file_mkdir_all", "_file_open"},
}
`
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	changed, err := FormatPath(path, Config{Write: true, Backup: true})
	if err != nil {
		t.Fatalf("FormatPath: %v", err)
	}
	if !changed {
		t.Fatal("expected change")
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "\t\t\"_file_close\",\n") {
		t.Fatalf("restyled file not written:\n%s", data)
	}
}

func TestFormatGoDoesNotRestyleMapLookingRawString(t *testing.T) {
	src := "package p\n\nvar specimen = `var policy = map[string][]string{\n\t\"lib/file/file.ai\": {\"_file_close\", \"_file_copy\", \"_file_delete\", \"_file_exists\", \"_file_list\", \"_file_mkdir\", \"_file_mkdir_all\", \"_file_open\"},\n}`\n"
	out, err := formatGo("specimen.go", src)
	if err != nil {
		t.Fatalf("formatGo: %v", err)
	}
	if !strings.Contains(out, `"lib/file/file.ai": {"_file_close", "_file_copy"`) {
		t.Fatalf("raw string contents were restyled:\n%s", out)
	}
}
