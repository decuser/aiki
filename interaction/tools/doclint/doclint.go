package doclint

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	Root          string
	Roots         []string
	Exclude       []string
	RequireHeader bool
	ValidTags     map[string]bool
}

type Violation struct {
	Path    string
	Line    int
	Message string
}

var tagRe = regexp.MustCompile(`^(NOW|PLAN|HIST|WHY|PHIL|RULE)\b`)

func findConfigRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	// collapse patterns like ./...
	if strings.Contains(dir, "...") {
		dir, _ = filepath.Abs(".")
	}

	for {
		probe := filepath.Join(dir, "doclint.ini")
		if _, err := os.Stat(probe); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no doclint.ini found starting from %s", start)
		}
		dir = parent
	}
}

func LoadConfig(path string) (*Config, error) {
	cfgRoot, err := findConfigRoot(path)
	if err != nil {
		return nil, err
	}

	iniPath := filepath.Join(cfgRoot, "doclint.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, fmt.Errorf("missing doclint.ini in %s", cfgRoot)
	}

	ini := parseIni(string(data))

	roots := splitList(ini.get("scope", "roots"))
	if len(roots) == 0 {
		return nil, fmt.Errorf("doclint.ini: scope.roots is required")
	}

	exclude := splitList(ini.get("scope", "exclude"))

	requireHeader := true
	rh := strings.TrimSpace(ini.get("contracts", "require_header"))
	if rh != "" {
		requireHeader = parseBool(rh)
	}

	valid := splitList(ini.get("tags", "valid"))
	if len(valid) == 0 {
		return nil, fmt.Errorf("doclint.ini: tags.valid is required")
	}
	validTags := make(map[string]bool)
	for _, t := range valid {
		validTags[t] = true
	}

	return &Config{
		Root:          cfgRoot,
		Roots:         roots,
		Exclude:       exclude,
		RequireHeader: requireHeader,
		ValidTags:     validTags,
	}, nil

}

func Check(cfg *Config) ([]Violation, error) {
	var violations []Violation

	for _, r := range cfg.Roots {
		rootPath := filepath.Join(cfg.Root, r)
		info, err := os.Stat(rootPath)
		if err != nil {
			return nil, fmt.Errorf("scope root not found: %s", r)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("scope root is not a directory: %s", r)
		}

		err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
					return filepath.SkipDir
				}
				return nil
			}

			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			rel, err := filepath.Rel(cfg.Root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			if isExcluded(rel, cfg.Exclude) {
				return nil
			}

			fileViolations, err := checkFile(cfg, path, rel)
			if err != nil {
				return err
			}
			violations = append(violations, fileViolations...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return violations, nil
}

func checkFile(cfg *Config, absPath string, relPath string) ([]Violation, error) {
	f, err := os.Open(absPath)
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

		// Header is a file level property.
		// If there is no header, do not emit a per line error here.
		// We will emit exactly one missing header error at end of file.
		if cfg.RequireHeader && !hasContract {
			continue
		}

		if !cfg.ValidTags[tag] {
			violations = append(violations, Violation{
				Path:    relPath,
				Line:    line,
				Message: "invalid tag: " + tag,
			})
			continue
		}

		if len(allowed) == 0 {
			violations = append(violations, Violation{
				Path:    relPath,
				Line:    line,
				Message: "contract header missing allowed list",
			})
			continue
		}

		if !allowed[tag] {
			violations = append(violations, Violation{
				Path:    relPath,
				Line:    line,
				Message: "tag not allowed by contract: " + tag,
			})
			continue
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	// Single file level checks

	if cfg.RequireHeader && !hasContract {
		violations = append(violations, Violation{
			Path:    relPath,
			Line:    1,
			Message: "missing contract header",
		})
	}

	if cfg.RequireHeader && hasContract && len(allowed) == 0 {
		violations = append(violations, Violation{
			Path:    relPath,
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

	parts := splitList(rest)
	for _, p := range parts {
		allowed[p] = true
	}
}

func isExcluded(rel string, patterns []string) bool {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = filepath.ToSlash(p)

		if rel == p {
			return true
		}

		if strings.HasPrefix(p, "**/") {
			suf := strings.TrimPrefix(p, "**/")
			if strings.HasSuffix(rel, suf) {
				return true
			}
		}

		if strings.HasSuffix(p, "/**") {
			prefix := strings.TrimSuffix(p, "/**")
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				return true
			}
		}
	}
	return false
}

func splitList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, ",", " ")
	fields := strings.Fields(s)
	var out []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

type iniFile map[string]map[string]string

func (i iniFile) get(section string, key string) string {
	sec := i[strings.ToLower(section)]
	if sec == nil {
		return ""
	}
	return sec[strings.ToLower(key)]
}

func parseIni(data string) iniFile {
	out := make(iniFile)
	section := "default"
	out[section] = make(map[string]string)

	sc := bufio.NewScanner(strings.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			if out[section] == nil {
				out[section] = make(map[string]string)
			}
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(line[:eq]))
		v := strings.TrimSpace(line[eq+1:])
		out[section][k] = v
	}
	return out
}
