package treecheck

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Finding struct {
	Path   string
	Reason string
}

type Result struct {
	Files     int
	Justified int
	Allowed   int
	Errors    []Finding
	Orphans   []Finding
}

type Checker struct {
	Root      string
	AllowFile string

	files     map[string]bool
	justified map[string]string
	allowed   map[string]string
	allow     []allowRule
}

type allowRule struct {
	raw    string
	prefix string
	glob   string
}

var packageRE = regexp.MustCompile(`(?m)^\s*package\s+"([^"]+)"`)
var importRE = regexp.MustCompile(`\b(?:import|use)\("([^"]+)"`)

func Check(root, allowFile string) (Result, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Result{}, err
	}
	c := &Checker{
		Root: abs, AllowFile: allowFile,
		files:     make(map[string]bool),
		justified: make(map[string]string),
		allowed:   make(map[string]string),
	}
	if err := c.loadFiles(); err != nil {
		return Result{}, err
	}
	if err := c.loadAllow(); err != nil {
		return Result{}, err
	}
	c.applyAllow()
	c.seedBuiltins()
	c.checkStructuralPairs()
	c.resolveAikiPackageReferences()
	c.resolveTextReferences()
	c.resolveDirectoryReadmes()

	var result Result
	result.Files = len(c.files)
	result.Justified = len(c.justified)
	result.Allowed = len(c.allowed)
	result.Errors = append(result.Errors, c.structuralErrors()...)
	errorPaths := make(map[string]bool, len(result.Errors))
	for _, finding := range result.Errors {
		errorPaths[finding.Path] = true
	}

	for path := range c.files {
		if errorPaths[path] {
			continue
		}
		if _, ok := c.justified[path]; ok {
			continue
		}
		if _, ok := c.allowed[path]; ok {
			continue
		}
		result.Orphans = append(result.Orphans, Finding{Path: path, Reason: "no recognized distribution relationship"})
	}
	sortFindings(result.Errors)
	sortFindings(result.Orphans)
	return result, nil
}

func (c *Checker) loadFiles() error {
	cmd := exec.Command("git", "-C", c.Root, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	out, err := cmd.Output()
	if err == nil {
		for _, raw := range bytes.Split(out, []byte{0}) {
			if len(raw) == 0 {
				continue
			}
			p := filepath.ToSlash(string(raw))
			if p == "" {
				continue
			}
			// git ls-files reports tracked paths from the index even when the
			// working-tree file has been removed. Treecheck describes the current
			// distribution, so absent paths must not be treated as present files.
			if info, statErr := os.Stat(filepath.Join(c.Root, filepath.FromSlash(p))); statErr != nil || info.IsDir() {
				continue
			}
			c.files[p] = true
		}
		return nil
	}

	return filepath.WalkDir(c.Root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(c.Root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".git" || strings.HasPrefix(rel, ".git/") {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == "aiki" || strings.HasPrefix(rel, "profile-out/") {
			return nil
		}
		c.files[rel] = true
		return nil
	})
}

func (c *Checker) loadAllow() error {
	p := filepath.Join(c.Root, c.AllowFile)
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("reading allow file %s: %w", c.AllowFile, err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		s = filepath.ToSlash(strings.TrimPrefix(s, "./"))
		r := allowRule{raw: s}
		if strings.HasSuffix(s, "/") {
			r.prefix = s
		} else {
			r.glob = s
		}
		c.allow = append(c.allow, r)
	}
	return sc.Err()
}

func (c *Checker) applyAllow() {
	for p := range c.files {
		for _, r := range c.allow {
			matched := false
			if r.prefix != "" {
				matched = strings.HasPrefix(p, r.prefix)
			} else if r.glob != "" {
				matched, _ = filepath.Match(r.glob, p)
			}
			if matched {
				c.allowed[p] = r.raw
				break
			}
		}
	}
}

func (c *Checker) mark(path, reason string) {
	if c.files[path] {
		c.justified[path] = reason
	}
}

func (c *Checker) seedBuiltins() {
	roots := []string{".gitignore", "LICENSE", "README.md", "buglist.md", "docs/audit-findings.md", "go.mod", "go.sum", "makefile", c.AllowFile}
	for _, p := range roots {
		c.mark(p, "repository root artifact")
	}
	for p := range c.files {
		switch {
		case strings.HasSuffix(p, ".go"):
			c.mark(p, "Go implementation/test source")
		case filepath.Base(p) == ".gitignore":
			c.mark(p, "directory ignore policy")
		case strings.HasPrefix(p, "lib/") && strings.HasSuffix(p, "_test.ai"):
			c.mark(p, "Aiki-native library test")
		case p == "engine/runtime/prelude/prelude_test.ai":
			c.mark(p, "Aiki-native prelude test")
		case (strings.HasPrefix(p, "test/behavior/") || strings.HasPrefix(p, "test/visual/")) && strings.HasSuffix(p, "_smoke.ai"):
			c.mark(p, "smoke-test specimen")
		case strings.HasPrefix(p, "test/structure/engine/") && strings.HasSuffix(p, "_engine.ai"):
			c.mark(p, "engine structural specimen")
		case strings.HasPrefix(p, "test/nativeffi/") && strings.HasSuffix(p, "_contract.ai"):
			c.mark(p, "native/FFI semantic conformance specimen")
		case strings.HasPrefix(p, "docs/releases/") && strings.HasSuffix(p, ".md"):
			c.mark(p, "release note")
		case strings.HasPrefix(p, "extra/samples/") && strings.HasSuffix(p, ".ai"):
			c.mark(p, "sample program discovered by runsamples")
		case strings.HasPrefix(p, "extra/profiling/") && matchedProfileDriver(p):
			c.mark(p, "profiling sweep driver")
		case p == "extra/profiling/sweep.sh":
			c.mark(p, "profiling sweep runner")
		case p == "engine/syntax/grammar.ebnfx" || p == "engine/syntax/grammar.help" || p == "engine/syntax/grammar.ebnfx.ebnf.gold":
			c.mark(p, "authoritative grammar artifact")
		case p == "engine/runtime/prelude/prelude.ai" || p == "engine/runtime/prelude/prelude.help" || p == "engine/runtime/prelude/prelude.doc":
			c.mark(p, "prelude artifact")
		case strings.HasPrefix(p, "selfhost/") && strings.HasSuffix(p, ".ai"):
			c.mark(p, "self-host implementation source")
		case p == "extra/editors/xed/aiki.lang":
			c.mark(p, "Xed GtkSourceView language definition")
		case p == "extra/editors/xed/aiki_lsp.plugin":
			c.mark(p, "Xed Aiki LSP plugin descriptor")
		case strings.HasPrefix(p, "extra/editors/xed/aiki_lsp/") && strings.HasSuffix(p, ".py"):
			c.mark(p, "Xed Aiki LSP plugin source")
		case p == "extra/editors/vscode/package.json":
			c.mark(p, "VS Code Aiki extension manifest")
		case p == "extra/editors/vscode/extension.js":
			c.mark(p, "VS Code Aiki LSP client source")
		case p == "extra/editors/vscode/language-configuration.json":
			c.mark(p, "VS Code Aiki editor configuration")
		case p == "extra/editors/vscode/syntaxes/aiki.tmLanguage.json":
			c.mark(p, "VS Code Aiki lexical grammar")
		case p == "experiments/README.md":
			c.mark(p, "experiments collection guide")
		case matchedExperimentArtifact(p):
			c.mark(p, "distributed experiment artifact")
		}
	}
}

func matchedExperimentArtifact(p string) bool {
	parts := strings.Split(filepath.ToSlash(p), "/")
	if len(parts) < 3 || parts[0] != "experiments" {
		return false
	}
	return experimentDirectoryRE.MatchString(parts[1])
}

var experimentDirectoryRE = regexp.MustCompile(`^[0-9]{3}-[a-z0-9]+(?:-[a-z0-9]+)*$`)

func matchedProfileDriver(p string) bool {
	base := filepath.Base(p)
	if len(base) < len("00-.ai") || !strings.HasSuffix(base, ".ai") {
		return false
	}
	return base[0] >= '0' && base[0] <= '9' && base[1] >= '0' && base[1] <= '9' && base[2] == '-'
}

func (c *Checker) checkStructuralPairs() {
	for p := range c.files {
		if strings.HasPrefix(p, "lib/") && strings.HasSuffix(p, ".ai") && !strings.HasSuffix(p, "_test.ai") {
			stem := strings.TrimSuffix(p, ".ai")
			c.mark(p, "standard-library module")
			if c.files[stem+".help"] {
				c.mark(stem+".help", "module help")
			}
			if c.files[stem+".doc"] {
				c.mark(stem+".doc", "module documentation")
			}
		}
		if (strings.HasPrefix(p, "test/behavior/") || strings.HasPrefix(p, "test/visual/")) && strings.HasSuffix(p, "_smoke.ai") {
			gold := strings.TrimSuffix(p, ".ai") + ".gold"
			if c.files[gold] {
				c.mark(gold, "smoke-test gold")
			}
		}
		if strings.HasPrefix(p, "test/structure/engine/") && strings.HasSuffix(p, "_engine.ai") {
			for _, stage := range []string{"lex", "parse", "eval"} {
				gold := p + "." + stage + ".gold"
				if c.files[gold] {
					c.mark(gold, "engine structural gold")
				}
			}
		}
		if strings.HasPrefix(p, "test/conformance/syntax/") && strings.HasSuffix(p, ".input") {
			projection := strings.TrimSuffix(p, ".input") + ".tokens"
			c.mark(p, "syntax conformance input")
			if c.files[projection] {
				c.mark(projection, "syntax conformance token projection")
			}
		}
		if p == "selfhost/token_authority.ai" && c.files["selfhost/token_authority.gold"] {
			c.mark("selfhost/token_authority.gold", "self-host authority projection")
		}
		if p == "extra/editors/xed/aiki_lsp.plugin" && c.files["extra/editors/xed/aiki_lsp/__init__.py"] {
			c.mark("extra/editors/xed/aiki_lsp.plugin", "Xed plugin descriptor paired with module")
			c.mark("extra/editors/xed/aiki_lsp/__init__.py", "Xed plugin module paired with descriptor")
		}
		if p == "extra/editors/vscode/package.json" {
			for _, q := range []string{
				"extra/editors/vscode/extension.js",
				"extra/editors/vscode/language-configuration.json",
				"extra/editors/vscode/syntaxes/aiki.tmLanguage.json",
			} {
				if c.files[q] {
					c.mark(q, "VS Code extension companion")
				}
			}
		}
	}
}

func isNamedPackageSource(path string) bool {
	if !strings.HasSuffix(path, ".ai") {
		return false
	}
	return strings.HasPrefix(path, "lib/") || strings.HasPrefix(path, "vendor/")
}

func (c *Checker) structuralErrors() []Finding {
	var out []Finding
	packages := make(map[string][]string)
	experimentDirs := make(map[string]bool)
	for p := range c.files {
		parts := strings.Split(filepath.ToSlash(p), "/")
		if len(parts) >= 3 && parts[0] == "experiments" && experimentDirectoryRE.MatchString(parts[1]) {
			experimentDirs["experiments/"+parts[1]] = true
		}
	}
	for dir := range experimentDirs {
		for _, companion := range []string{"README.md", "experiment/PROCEDURE.md", "experiment/run.sh"} {
			path := dir + "/" + companion
			if !c.files[path] {
				out = append(out, Finding{Path: dir, Reason: "experiment missing " + companion})
			}
		}
	}

	for p := range c.files {
		if strings.HasPrefix(p, "lib/") && strings.HasSuffix(p, ".ai") && !strings.HasSuffix(p, "_test.ai") {
			stem := strings.TrimSuffix(p, ".ai")
			for _, ext := range []string{".help", ".doc"} {
				q := stem + ext
				if !c.files[q] {
					out = append(out, Finding{Path: p, Reason: "module missing " + filepath.Base(q)})
				}
			}
		}
		if strings.HasPrefix(p, "lib/") && (strings.HasSuffix(p, ".help") || strings.HasSuffix(p, ".doc")) {
			stem := strings.TrimSuffix(strings.TrimSuffix(p, ".help"), ".doc")
			if !c.files[stem+".ai"] {
				out = append(out, Finding{Path: p, Reason: "module companion has no .ai owner"})
			}
		}
		if (strings.HasPrefix(p, "test/behavior/") || strings.HasPrefix(p, "test/visual/")) && strings.HasSuffix(p, "_smoke.ai") {
			gold := strings.TrimSuffix(p, ".ai") + ".gold"
			if !c.files[gold] {
				out = append(out, Finding{Path: p, Reason: "smoke specimen missing .gold"})
			}
		}
		if (strings.HasPrefix(p, "test/behavior/") || strings.HasPrefix(p, "test/visual/")) && strings.HasSuffix(p, ".gold") {
			ai := strings.TrimSuffix(p, ".gold") + ".ai"
			if !c.files[ai] {
				out = append(out, Finding{Path: p, Reason: "smoke gold has no .ai specimen"})
			}
		}
		if strings.HasPrefix(p, "test/structure/engine/") && strings.Contains(p, "_engine.ai.") && strings.HasSuffix(p, ".gold") {
			i := strings.Index(p, ".ai.")
			ai := p[:i+3]
			if !c.files[ai] {
				out = append(out, Finding{Path: p, Reason: "engine gold has no .ai specimen"})
			}
		}
		if strings.HasPrefix(p, "test/structure/engine/") && strings.HasSuffix(p, "_engine.ai") {
			for _, stage := range []string{"lex", "parse", "eval"} {
				gold := p + "." + stage + ".gold"
				if !c.files[gold] {
					out = append(out, Finding{Path: p, Reason: "engine specimen missing " + stage + " gold"})
				}
			}
		}
		if strings.HasPrefix(p, "test/conformance/syntax/") && strings.HasSuffix(p, ".input") {
			projection := strings.TrimSuffix(p, ".input") + ".tokens"
			if !c.files[projection] {
				out = append(out, Finding{Path: p, Reason: "syntax conformance input missing .tokens projection"})
			}
		}
		if strings.HasPrefix(p, "test/conformance/syntax/") && strings.HasSuffix(p, ".tokens") {
			input := strings.TrimSuffix(p, ".tokens") + ".input"
			if !c.files[input] {
				out = append(out, Finding{Path: p, Reason: "syntax conformance projection has no .input specimen"})
			}
		}
		if p == "selfhost/token_authority.gold" && !c.files["selfhost/token_authority.ai"] {
			out = append(out, Finding{Path: p, Reason: "self-host authority projection has no .ai owner"})
		}
		if p == "extra/editors/xed/aiki_lsp.plugin" && !c.files["extra/editors/xed/aiki_lsp/__init__.py"] {
			out = append(out, Finding{Path: p, Reason: "Xed plugin descriptor has no Python module"})
		}
		if p == "extra/editors/xed/aiki_lsp/__init__.py" && !c.files["extra/editors/xed/aiki_lsp.plugin"] {
			out = append(out, Finding{Path: p, Reason: "Xed plugin module has no descriptor"})
		}
		if p == "extra/editors/vscode/package.json" {
			for _, q := range []string{
				"extra/editors/vscode/extension.js",
				"extra/editors/vscode/language-configuration.json",
				"extra/editors/vscode/syntaxes/aiki.tmLanguage.json",
			} {
				if !c.files[q] {
					out = append(out, Finding{Path: p, Reason: "VS Code extension missing " + strings.TrimPrefix(q, "extra/editors/vscode/")})
				}
			}
		}
		if strings.HasPrefix(p, "extra/editors/vscode/") && p != "extra/editors/vscode/README.md" && p != "extra/editors/vscode/package.json" {
			if !c.files["extra/editors/vscode/package.json"] {
				out = append(out, Finding{Path: p, Reason: "VS Code extension companion has no package.json owner"})
			}
		}
		if isNamedPackageSource(p) {
			data, err := os.ReadFile(filepath.Join(c.Root, filepath.FromSlash(p)))
			if err == nil {
				if m := packageRE.FindSubmatch(data); m != nil {
					packages[string(m[1])] = append(packages[string(m[1])], p)
				}
			}
		}
	}
	for name, paths := range packages {
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		for _, p := range paths {
			out = append(out, Finding{Path: p, Reason: fmt.Sprintf("duplicate Aiki package %q (%s)", name, strings.Join(paths, ", "))})
		}
	}
	return out
}

func (c *Checker) resolveAikiPackageReferences() {
	packageFiles := make(map[string][]string)
	for p := range c.files {
		if !isNamedPackageSource(p) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(c.Root, filepath.FromSlash(p)))
		if err != nil {
			continue
		}
		if m := packageRE.FindSubmatch(data); m != nil {
			packageFiles[string(m[1])] = append(packageFiles[string(m[1])], p)
		}
	}

	changed := true
	for changed {
		changed = false
		for p := range c.justified {
			if !strings.HasSuffix(p, ".ai") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(c.Root, filepath.FromSlash(p)))
			if err != nil {
				continue
			}
			for _, m := range importRE.FindAllSubmatch(data, -1) {
				name := string(m[1])
				if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
					target := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(p), name)))
					if filepath.Ext(target) == "" {
						target += ".ai"
					}
					if c.files[target] {
						if _, ok := c.justified[target]; !ok {
							c.mark(target, "imported by "+p)
							changed = true
						}
					}
					continue
				}
				for _, target := range packageFiles[name] {
					if _, ok := c.justified[target]; ok {
						continue
					}
					c.mark(target, "imported by "+p)
					changed = true
				}
			}
		}
	}
}

func (c *Checker) resolveTextReferences() {
	changed := true
	for changed {
		changed = false
		var sources []string
		for p := range c.justified {
			if isTextPath(p) {
				sources = append(sources, p)
			}
		}
		sort.Strings(sources)
		for _, src := range sources {
			data, err := os.ReadFile(filepath.Join(c.Root, filepath.FromSlash(src)))
			if err != nil {
				continue
			}
			text := string(data)
			for target := range c.files {
				if target == src {
					continue
				}
				if _, ok := c.justified[target]; ok {
					continue
				}
				if strings.Contains(text, target) {
					c.mark(target, "referenced by "+src)
					changed = true
				}
			}
		}
	}
}

func (c *Checker) resolveDirectoryReadmes() {
	changed := true
	for changed {
		changed = false
		for p := range c.files {
			if filepath.Base(p) != "README.md" {
				continue
			}
			if _, ok := c.justified[p]; ok {
				continue
			}
			dir := filepath.ToSlash(filepath.Dir(p))
			prefix := dir + "/"
			for q := range c.justified {
				if q != p && strings.HasPrefix(q, prefix) {
					c.mark(p, "README for active distribution subtree")
					changed = true
					break
				}
			}
		}
	}
}

func isTextPath(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".go", ".ai", ".md", ".help", ".doc", ".sh", ".lang", ".ebnfx":
		return true
	}
	return filepath.Base(p) == "makefile" || strings.HasPrefix(p, "hooks/")
}

func sortFindings(v []Finding) {
	sort.Slice(v, func(i, j int) bool {
		if v[i].Path == v[j].Path {
			return v[i].Reason < v[j].Reason
		}
		return v[i].Path < v[j].Path
	})
}
