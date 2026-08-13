package lint

import (
	"os"
	"path/filepath"
	"strings"
)

func expandLintPaths(args []string) ([]string, error) {
	var files []string
	for _, a := range args {
		if strings.HasSuffix(a, "/...") || strings.HasSuffix(a, string(filepath.Separator)+"...") {
			dir := strings.TrimSuffix(a, "/...")
			dir = strings.TrimSuffix(dir, string(filepath.Separator)+"...")
			if dir == "" {
				dir = "."
			}
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
				if strings.HasSuffix(path, ".ai") && !strings.HasSuffix(path, "_test.ai") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		info, err := os.Stat(a)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err := filepath.Walk(a, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
					return filepath.SkipDir
				}
				if info.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".ai") && !strings.HasSuffix(path, "_test.ai") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		if strings.HasSuffix(a, ".ai") && !strings.HasSuffix(a, "_test.ai") {
			files = append(files, a)
		}
	}
	return files, nil
}
