package boundary_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type pkgJSON struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
}

func goListAll(t *testing.T) []pkgJSON {
	t.Helper()

	root := moduleDir(t)

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = root

	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("go list failed: %v\n%s", err, string(ee.Stderr))
		}
		t.Fatalf("go list failed: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []pkgJSON
	for {
		var p pkgJSON
		if err := dec.Decode(&p); err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("decode go list json: %v", err)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func anyHasPrefix(list []string, prefix string) bool {
	for _, x := range list {
		if strings.HasPrefix(x, prefix) {
			return true
		}
	}
	return false
}

func TestHostBoundaryImports(t *testing.T) {
	pkgs := goListAll(t)

	// Only these packages may import runtime hal directly.
	allowedToImportRuntimeHal := map[string]bool{
		"aiki/cmd/aiki":           true,
		"aiki/semantics/testutil": true,
		"aiki/runtime/hal":        true, // self import list does not include itself, but keep for clarity
	}

	runtimeHalImport := "aiki/runtime/hal"

	// Packages that must remain host free.
	hostFreePrefixes := []string{
		"aiki/semantics/",
		"aiki/syntax",
		"aiki/runtime/prelude",
	}

	// Host capability packages that must not appear in host free packages.
	forbiddenHostImports := map[string]bool{
		"os":            true,
		"os/exec":       true,
		"path/filepath": true,
		"net":           true,
		"net/http":      true,
		"time":          true,
		"syscall":       true,
		"unsafe":        true,
		"reflect":       true,
		"runtime/cgo":   true,

		// Third party host heavy dependencies
		"github.com/hajimehoshi/ebiten/v2": true,
	}

	var violations []string

	for _, p := range pkgs {
		imports := p.Imports

		if strings.Contains(p.ImportPath, "/testutil") {
			continue
		}

		// Rule 1: direct runtime hal import is restricted.
		if has(imports, runtimeHalImport) && !allowedToImportRuntimeHal[p.ImportPath] {
			violations = append(violations,
				fmt.Sprintf("%s imports %s", p.ImportPath, runtimeHalImport))
		}

		// Rule 2: host free packages must not import host capability packages.
		isHostFree := false
		for _, pref := range hostFreePrefixes {
			if p.ImportPath == pref || strings.HasPrefix(p.ImportPath, pref) {
				isHostFree = true
				break
			}
		}
		if isHostFree {
			for _, imp := range imports {
				if forbiddenHostImports[imp] {
					violations = append(violations,
						fmt.Sprintf("%s imports forbidden host package %s", p.ImportPath, imp))
				}
				// Catch ebiten subpackages too
				if strings.HasPrefix(imp, "github.com/hajimehoshi/ebiten/v2/") {
					violations = append(violations,
						fmt.Sprintf("%s imports forbidden host package %s", p.ImportPath, imp))
				}
			}
		}

		// Optional: tighten by forbidding any ebiten import outside runtime hal
		if p.ImportPath != "aiki/runtime/hal" && strings.HasPrefix(p.ImportPath, "aiki/") {
			if anyHasPrefix(imports, "github.com/hajimehoshi/ebiten/v2") {
				violations = append(violations,
					fmt.Sprintf("%s imports ebiten outside runtime hal", p.ImportPath))
			}
		}
	}

	if len(violations) > 0 {
		t.Fatalf("host boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func moduleDir(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go env GOMOD failed: %v", err)
	}
	gomod := strings.TrimSpace(string(out))
	if gomod == "" || gomod == "NUL" {
		t.Fatalf("GOMOD not set, not in a module")
	}
	return filepath.Dir(gomod)
}
