package lint

import (
	"os"
	"path/filepath"
	"strings"

	"aiki/cmd/internal/testfixture"
)

func expandLintPaths(args []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		if isRecursivePath(arg) {
			dir := recursiveRoot(arg)
			if err := appendLintDir(&files, dir); err != nil {
				return nil, err
			}
			continue
		}

		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			if err := appendLintDir(&files, arg); err != nil {
				return nil, err
			}
			continue
		}

		include, err := lintSourceCandidate(arg)
		if err != nil {
			return nil, err
		}
		if include {
			files = append(files, arg)
		}
	}
	return files, nil
}

func isRecursivePath(path string) bool {
	return strings.HasSuffix(path, "/...") || strings.HasSuffix(path, string(filepath.Separator)+"...")
}

func recursiveRoot(path string) string {
	dir := strings.TrimSuffix(path, "/...")
	dir = strings.TrimSuffix(dir, string(filepath.Separator)+"...")
	if dir == "" {
		return "."
	}
	return dir
}

func appendLintDir(files *[]string, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		include, err := lintSourceCandidate(path)
		if err != nil {
			return err
		}
		if include {
			*files = append(*files, path)
		}
		return nil
	})
}

// lintSourceCandidate is the single disposition point for files entering the
// linter's AST traversal. Parse-negative smoke fixtures are intentionally
// excluded; ordinary source, including malformed source, remains lint input.
func lintSourceCandidate(path string) (bool, error) {
	if !strings.HasSuffix(path, ".ai") || strings.HasSuffix(path, "_test.ai") {
		return false, nil
	}
	skip, err := testfixture.IsParseNegative(path)
	if err != nil {
		return false, err
	}
	return !skip, nil
}
