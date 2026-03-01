package doclint

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Options struct {
	RequireHeader bool
}

type Violation struct {
	Path    string
	Line    int
	Message string
}

// tagRe matches uppercase words at start of line (potential tags)
var tagRe = regexp.MustCompile(`^([A-Z]{2,})\b`)

// Check lints markdown files at the given paths.
// Paths can be files, directories, or "./..." for recursive.
func Check(paths []string, opts Options) ([]Violation, error) {
	var violations []Violation

	files, err := expandPaths(paths)
	if err != nil {
		return nil, err
	}

	for _, path := range files {
		fileViolations, err := checkFile(path, opts)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}

	return violations, nil
}

// expandPaths resolves paths to a list of .md files.
func expandPaths(paths []string) ([]string, error) {
	var files []string

	for _, p := range paths {
		// Handle ./... recursive pattern
		if strings.HasSuffix(p, "/...") || p == "./..." {
			root := strings.TrimSuffix(p, "/...")
			if root == "." || root == "" {
				root = "."
			}
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
						return filepath.SkipDir
					}
					return nil
				}
				if strings.HasSuffix(path, ".md") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
						return filepath.SkipDir
					}
					return nil
				}
				if strings.HasSuffix(path, ".md") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
	}

	return files, nil
}

func checkFile(path string, opts Options) ([]Violation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	line := 0

	hasContract := false
	allowed := make(map[string]bool)

	inContract := false
	contractDone := false

	var violations []Violation

	for sc.Scan() {
		line++
		txt := sc.Text()

		// contract header parsing
		if !contractDone {
			trim := strings.TrimSpace(txt)

			if !inContract && strings.HasPrefix(trim, "<!--") && strings.Contains(trim, "contract") {
				inContract = true
				hasContract = true
			}

			if inContract {
				if strings.Contains(trim, "allowed:") {
					parseAllowedLine(trim, allowed)
				}
				if strings.Contains(trim, "-->") {
					inContract = false
					contractDone = true
				}
				continue
			}
		}

		m := tagRe.FindStringSubmatch(txt)
		if m == nil {
			continue
		}
		tag := m[1]

		// No header means we can't validate tags
		if opts.RequireHeader && !hasContract {
			continue
		}

		// If file has no contract, skip tag validation
		if !hasContract {
			continue
		}

		if len(allowed) == 0 {
			violations = append(violations, Violation{
				Path:    path,
				Line:    line,
				Message: "contract header missing allowed list",
			})
			continue
		}

		if !allowed[tag] {
			violations = append(violations, Violation{
				Path:    path,
				Line:    line,
				Message: "tag not allowed by contract: " + tag,
			})
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	// File-level checks
	if opts.RequireHeader && !hasContract {
		// Skip files without contract headers
		return nil, nil
	}

	if hasContract && len(allowed) == 0 {
		violations = append(violations, Violation{
			Path:    path,
			Line:    1,
			Message: "contract header missing allowed list",
		})
	}

	return violations, nil
}

func parseAllowedLine(line string, allowed map[string]bool) {
	i := strings.Index(line, "allowed:")
	if i < 0 {
		return
	}
	rest := strings.TrimSpace(line[i+len("allowed:"):])
	rest = strings.TrimSuffix(rest, "-->")
	rest = strings.TrimSpace(rest)

	// Split on comma or space
	rest = strings.ReplaceAll(rest, ",", " ")
	for _, t := range strings.Fields(rest) {
		t = strings.TrimSpace(t)
		if t != "" {
			allowed[t] = true
		}
	}
}
