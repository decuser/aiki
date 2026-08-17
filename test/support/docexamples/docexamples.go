package docexamples

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"aiki/engine/runtime/modules"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

var reExpect = regexp.MustCompile(`^(.+?)\s{2,}#\s+(.+)$`)
var reValue = regexp.MustCompile(`^(?:-?\d+(?:/\d+)?|".*"|:\w+|true|false|\[.*\]|'.'|<bytes:\d+.*>|<(?:module|channel|file|fn)\b.*>|@\w+)$`)

type Example struct {
	Module    string
	Name      string
	Preamble  string
	Code      []string
	Expects   []string
	Unchecked bool
}

func Load(root string) ([]Example, error) {
	type docFile struct{ name, path string }
	var files []docFile

	preludePath := filepath.Join(root, "engine", "runtime", "prelude", "prelude.doc")
	if _, err := os.Stat(preludePath); err == nil {
		files = append(files, docFile{name: "prelude", path: preludePath})
	}

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, fmt.Errorf("loading grammar: %w", err)
	}
	var roots []string
	for _, r := range modules.DistributionModuleRoots() {
		roots = append(roots, filepath.Join(root, r))
	}
	registry := modules.NewModuleRegistry(roots)
	if err := registry.Scan(g); err != nil {
		return nil, fmt.Errorf("scanning distribution modules: %w", err)
	}
	for _, name := range registry.ListCanonicalPackages() {
		aiPath, ok := registry.Lookup(name)
		if !ok {
			return nil, fmt.Errorf("registry listed %s but cannot locate it", name)
		}
		docPath := strings.TrimSuffix(aiPath, ".ai") + ".doc"
		if _, err := os.Stat(docPath); err == nil {
			files = append(files, docFile{name: name, path: docPath})
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	var out []Example
	for _, file := range files {
		examples, err := ParseFile(file.name, file.path)
		if err != nil {
			return nil, err
		}
		out = append(out, examples...)
	}
	return out, nil
}

func ParseFile(moduleName, path string) ([]Example, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot read doc: %w", moduleName, err)
	}
	var preambleLines, bodyLines []string
	for _, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(line, "@preamble ") {
			preambleLines = append(preambleLines, strings.TrimPrefix(line, "@preamble "))
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	preamble := strings.Join(preambleLines, "\n")
	body := strings.Join(bodyLines, "\n")

	var out []Example
	for _, entry := range strings.Split(body, "\n===\n") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		lines := strings.Split(entry, "\n")
		name := strings.TrimSpace(lines[0])
		if name == "" {
			continue
		}
		ex := Example{Module: moduleName, Name: name, Preamble: preamble}
		for _, line := range lines[1:] {
			trimmed := strings.TrimSpace(line)
			if trimmed == "@unchecked" {
				ex.Unchecked = true
				continue
			}
			if trimmed == "" || trimmed == "Example" {
				continue
			}
			if m := reExpect.FindStringSubmatch(line); m != nil {
				expr := strings.TrimSpace(m[1])
				expected := strings.TrimSpace(m[2])
				if reValue.MatchString(expected) {
					ex.Code = append(ex.Code, "println(inspect("+expr+"))")
					ex.Expects = append(ex.Expects, expected)
					continue
				}
			}
			if strings.HasPrefix(trimmed, "let ") || strings.HasPrefix(trimmed, "use(") || strings.HasPrefix(trimmed, "import(") {
				ex.Code = append(ex.Code, line)
			}
		}
		out = append(out, ex)
	}
	return out, nil
}
