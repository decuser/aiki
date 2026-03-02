package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	aikifmt "aiki/cmd/subcommands/tools/fmt"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// CheckFormatting returns the list of files that are not already formatted.
// It does not rewrite files.
func CheckFormatting(args []string) ([]string, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, err
	}
	var bad []string
	for _, a := range args {
		if strings.HasSuffix(a, "/...") || strings.HasSuffix(a, string(filepath.Separator)+"...") {
			dir := strings.TrimSuffix(a, "/...")
			dir = strings.TrimSuffix(dir, string(filepath.Separator)+"...")
			if dir == "" {
				dir = "."
			}
			b, err := checkDir(g, dir)
			if err != nil {
				return nil, err
			}
			bad = append(bad, b...)
			continue
		}
		ok, err := checkFile(g, a)
		if err != nil {
			return nil, err
		}
		if !ok {
			bad = append(bad, a)
		}
	}
	return bad, nil
}

func checkDir(g *grammar.Grammar, dir string) ([]string, error) {
	var bad []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".ai") {
			return nil
		}
		ok, err := checkFile(g, path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if !ok {
			bad = append(bad, path)
		}
		return nil
	})
	return bad, err
}

func checkFile(g *grammar.Grammar, path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	formatted, err := aikifmt.FormatSource(g, path, string(data))
	if err != nil {
		return false, err
	}
	return formatted == string(data), nil
}
