package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "aiki/hal/canvas"
	"aiki/lang/eval"
	"aiki/lang/parser"
	"aiki/strict"
)

// Lint checks a file or directory. Returns count of files checked.
func Lint(pathStr string) (int, error) {
	var allowed []string
	for k := range eval.HAL {
		allowed = append(allowed, k)
	}
	allowed = append(allowed, strict.Exports()...)

	if strings.HasSuffix(pathStr, "/...") {
		dir := strings.TrimSuffix(pathStr, "/...")
		if dir == "." || dir == "" {
			dir = "."
		}
		return lintDir(dir, allowed)
	}

	if err := lintFile(pathStr, allowed); err != nil {
		return 0, err
	}
	return 1, nil
}

func lintDir(dir string, allowed []string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".ai") {
			if err := lintFile(path, allowed); err != nil {
			    fmt.Fprintf(os.Stderr, "%s\n", err)
			} else {
			    count++
			}
		}
		return nil
	})
	return count, err
}

func lintFile(path string, allowed []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	p := parser.New(string(data))
	program := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse error: %s", p.Errors()[0])
	}

	issues := Check(program, allowed)
	if len(issues) > 0 {
		fmt.Println(path)
		for _, issue := range issues {
			fmt.Printf("\t%s\n", issue)
		}
	}

	return nil
}
