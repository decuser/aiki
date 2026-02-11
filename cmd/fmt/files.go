package fmt

import (
	gofmt "fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/lang/parser"
)

// Format formats a file or directory. Returns count of files formatted.
func Format(pathStr string) (int, error) {
	if strings.HasSuffix(pathStr, "/...") {
		dir := strings.TrimSuffix(pathStr, "/...")
		if dir == "." || dir == "" {
			dir = "."
		}
		return formatDir(dir)
	}

	if err := formatFile(pathStr); err != nil {
		return 0, err
	}
	return 1, nil
}

func formatDir(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(path, ".ai") {
			if err := formatFile(path); err != nil {
				gofmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			} else {
				count++
			}
		}
		return nil
	})
	return count, err
}

func formatFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	p := parser.New(string(data))
	program := p.Parse()
	if len(p.Errors()) > 0 {
		return gofmt.Errorf("parse error: %s", p.Errors()[0])
	}

	formatted := PrintWithComments(program, p.Comments())

	if formatted != string(data) {
		gofmt.Println(path)
		if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
			return err
		}
	}

	return nil
}
