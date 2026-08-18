package invariant

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var ambientOSNames = map[string]bool{
	"Stdin": true, "Stdout": true, "Stderr": true, "Args": true,
	"Getenv": true, "LookupEnv": true, "Environ": true, "Setenv": true, "Unsetenv": true, "Getwd": true, "Chdir": true,
}

func TestAikiFacingIOAndSystemUseRuntimeOwnedState(t *testing.T) {
	if err := validateNoAmbientRuntimeGlobals(loadRuntimeOwnedSurface(t)); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeOwnershipInvariantRejectsAmbientGlobalRegression(t *testing.T) {
	sources := loadRuntimeOwnedSurface(t)
	sources["engine/runtime/hal/substrate/forbidden.go"] = "package substrate\nimport \"os\"\nvar forbiddenRegression = os.Args\n"
	err := validateNoAmbientRuntimeGlobals(sources)
	if err == nil || !strings.Contains(err.Error(), "os.Args") {
		t.Fatalf("expected ambient-global invariant failure, got %v", err)
	}
}

func loadRuntimeOwnedSurface(t *testing.T) map[string]string {
	t.Helper()
	root := distributionRoot(t)
	patterns := []string{
		"engine/runtime/hal/substrate/builtins_io*.go",
		"engine/runtime/hal/substrate/builtins_system.go",
		"engine/runtime/hal/substrate/builtins_workdir.go",
		"engine/runtime/hal/substrate/builtins_intrinsic.go",
		"engine/runtime/hal/substrate/builtins_process.go",
		"engine/runtime/hal/substrate/builtins_signal.go",
		"engine/runtime/hal/substrate/builtins_network.go",
		"engine/runtime/hal/substrate/builtins_terminal.go",
	}
	out := map[string]string{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			rel, _ := filepath.Rel(root, path)
			out[filepath.ToSlash(rel)] = string(data)
		}
	}
	return out
}

func validateNoAmbientRuntimeGlobals(sources map[string]string) error {
	var problems []string
	for path, source := range sources {
		f, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if ok && id.Name == "os" && ambientOSNames[sel.Sel.Name] {
				problems = append(problems, fmt.Sprintf("%s uses ambient process state os.%s", path, sel.Sel.Name))
			}
			return true
		})
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("runtime-ownership invariant failure:\n  %s", strings.Join(problems, "\n  "))
}
