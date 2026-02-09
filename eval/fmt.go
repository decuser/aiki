package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/ast"
	"aiki/parser"
	"aiki/value"
)

func init() {
	builtins["fmt"] = &value.Builtin{
		Name: "fmt",
		Fn:   builtinFmt,
	}
}

// Format formats a file or directory. Exported for command line use.
func Format(pathStr string) value.Value {
	// Check for recursive pattern
	if strings.HasSuffix(pathStr, "/...") {
		dir := strings.TrimSuffix(pathStr, "/...")
		if dir == "." || dir == "" {
			dir = "."
		}
		count, err := formatDir(dir)
		if err != nil {
			return value.NewError("fmt: %s", err)
		}
		return value.NewNumber(int64(count), 1)
	}

	// Single file
	if err := formatFile(pathStr); err != nil {
		return value.NewError("fmt: %s", err)
	}
	return value.NewNumber(1, 1)
}

func builtinFmt(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("fmt: want 1 argument, got %d", len(args))
	}

	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("fmt: expected string argument")
	}

	return Format(path.Value)
}

func formatDir(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		// Only process .ai files
		if !info.IsDir() && strings.HasSuffix(path, ".ai") {
			if err := formatFile(path); err != nil {
				// Print error but continue
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			} else {
				count++
			}
		}
		return nil
	})
	return count, err
}

func formatFile(path string) error {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse
	p := parser.New(string(data))
	program := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	// Format with comments
	formatted := ast.PrintWithComments(program, p.Comments())

	// Write back only if changed
	if formatted != string(data) {
		if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
			return err
		}
	}

	return nil
}
