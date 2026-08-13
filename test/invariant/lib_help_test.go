package invariant

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/help"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// The engine already enforces that prelude.ai and prelude.help agree. That
// invariant stops at the prelude: nothing checked the supplied library modules,
// which is how regex/ffi came to document every one of its procedures with the
// arguments reversed while the whole test suite stayed green.
//
// These tests extend the same relationship to the shipped modules. For each
// one they check that exports, help entries, and doc entries name the same
// procedures, and that each help template agrees with the procedure it
// describes in name, parameter count, and parameter order.
//
// Modules are discovered through the same registry import uses, restricted to
// the distribution roots. Adding a module anywhere the distribution ships one
// therefore brings it under these checks without touching this file.

var (
	reExport   = regexp.MustCompile(`(?s)export\(([^)]*)\)`)
	reSymbol   = regexp.MustCompile(`:(\w+)`)
	reLetFunc  = regexp.MustCompile(`(?m)^let\s+(\w+)\s*=\s*\(([^)]*)\)\s*\{`)
	reTemplate = regexp.MustCompile(`^(\w+)\(([^)]*)\)$`)
)

type module struct {
	name     string // e.g. "regex/ffi"
	aiPath   string
	exports  []string
	params   map[string][]string // exported name -> ordinary parameter names
	hasRest  map[string]bool
	helpPath string
	docPath  string
}

// distributionRoot locates the repository root relative to this test's package.
func distributionRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("cannot find distribution root at %s: %v", root, err)
	}
	return root
}

// shippedModulePaths returns the source path of every module the distribution
// ships, keyed by package name, discovered through the registry import uses.
func shippedModulePaths(t *testing.T) map[string]string {
	t.Helper()

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}

	root := distributionRoot(t)
	var roots []string
	for _, r := range substrate.DistributionModuleRoots() {
		roots = append(roots, filepath.Join(root, r))
	}

	registry := substrate.NewModuleRegistry(roots)
	if err := registry.Scan(g); err != nil {
		t.Fatalf("scanning distribution roots: %v", err)
	}

	paths := map[string]string{}
	for _, name := range registry.ListPackages() {
		path, ok := registry.Lookup(name)
		if !ok {
			t.Fatalf("registry listed %s but cannot locate it", name)
		}
		paths[name] = path
	}
	if len(paths) == 0 {
		t.Fatalf("no modules found under %v", roots)
	}
	return paths
}

func splitArgs(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// loadModules reads each shipped module, pairing its exports and parameter
// lists with the help and doc files that sit beside its source.
func loadModules(t *testing.T) []module {
	t.Helper()

	var mods []module
	for name, path := range shippedModulePaths(t) {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: cannot read %s: %v", name, path, err)
			continue
		}

		mod := module{
			name:    name,
			aiPath:  path,
			params:  map[string][]string{},
			hasRest: map[string]bool{},
		}
		if m := reExport.FindSubmatch(src); m != nil {
			for _, sym := range reSymbol.FindAllSubmatch(m[1], -1) {
				mod.exports = append(mod.exports, string(sym[1]))
			}
		}
		for _, f := range reLetFunc.FindAllSubmatch(src, -1) {
			fname := string(f[1])
			var ordinary []string
			for _, p := range splitArgs(string(f[2])) {
				if strings.HasPrefix(p, "...") {
					mod.hasRest[fname] = true
					continue
				}
				ordinary = append(ordinary, p)
			}
			mod.params[fname] = ordinary
		}

		base := strings.TrimSuffix(path, ".ai")
		mod.helpPath = base + ".help"
		mod.docPath = base + ".doc"
		mods = append(mods, mod)
	}

	sort.Slice(mods, func(i, j int) bool { return mods[i].name < mods[j].name })
	return mods
}

func helpEntries(t *testing.T, m module) map[string]help.FuncEntry {
	t.Helper()
	src, err := os.ReadFile(m.helpPath)
	if err != nil {
		t.Errorf("%s: cannot read %s: %v", m.name, filepath.Base(m.helpPath), err)
		return nil
	}
	entries, err := help.ParseHelpFile(m.helpPath, string(src))
	if err != nil {
		t.Errorf("%s: parsing help: %v", m.name, err)
		return nil
	}
	return entries
}

// TestLibHelpCoversExports checks that a module's help file describes exactly
// the procedures the module exports: no export without an entry, and no entry
// naming a procedure that does not exist.
func TestLibHelpCoversExports(t *testing.T) {
	for _, m := range loadModules(t) {
		entries := helpEntries(t, m)
		if entries == nil {
			continue
		}

		var missing []string
		for _, name := range m.exports {
			if _, ok := entries[name]; !ok {
				missing = append(missing, name)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("%s: exported but absent from help: %s", m.name, strings.Join(missing, ", "))
		}

		exported := map[string]bool{}
		for _, name := range m.exports {
			exported[name] = true
		}
		var phantom []string
		for name := range entries {
			if !exported[name] {
				phantom = append(phantom, name)
			}
		}
		sort.Strings(phantom)
		if len(phantom) > 0 {
			t.Errorf("%s: help describes procedures the module does not export: %s",
				m.name, strings.Join(phantom, ", "))
		}
	}
}

// TestLibHelpTemplatesMatchSignatures checks each @template against the
// procedure it describes. Names, parameter count, and parameter order must
// agree; a procedure with a rest parameter may be described with additional
// arguments beyond its ordinary ones.
func TestLibHelpTemplatesMatchSignatures(t *testing.T) {
	for _, m := range loadModules(t) {
		entries := helpEntries(t, m)
		if entries == nil {
			continue
		}

		for name, entry := range entries {
			params, defined := m.params[name]
			if !defined {
				continue // reported by TestLibHelpCoversExports
			}
			tm := reTemplate.FindStringSubmatch(strings.TrimSpace(entry.Template))
			if tm == nil {
				t.Errorf("%s.%s: template %q is not of the form name(args)", m.name, name, entry.Template)
				continue
			}
			if tm[1] != name {
				t.Errorf("%s.%s: template names %q", m.name, name, tm[1])
			}
			args := splitArgs(tm[2])

			if m.hasRest[name] {
				if len(args) < len(params) {
					t.Errorf("%s.%s: template has %d arguments, procedure requires %d",
						m.name, name, len(args), len(params))
					continue
				}
				args = args[:len(params)]
			} else if len(args) != len(params) {
				t.Errorf("%s.%s: template (%s) but procedure takes (%s)",
					m.name, name, strings.Join(args, ", "), strings.Join(params, ", "))
				continue
			}

			for i := range params {
				if args[i] != params[i] {
					t.Errorf("%s.%s: template (%s) but procedure takes (%s)",
						m.name, name, strings.Join(args, ", "), strings.Join(params, ", "))
					break
				}
			}
		}
	}
}

// TestLibDocFilesWellFormed checks that each module doc file is in the
// ===-separated form that ParseDocFile reads: an entry is a name on its own
// line followed by its text, and entries are separated by a line containing
// only ===. Two departures from that form are silent, so they are named here.
// stripPreambleLines removes @preamble directives from doc file text.
// These are consumed by TestDocExamplesExecutable and are not part of the
// entry namespace that ParseDocFile reads.
func stripPreambleLines(text string) string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "@preamble ") {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func TestLibDocFilesWellFormed(t *testing.T) {
	for _, m := range loadModules(t) {
		src, err := os.ReadFile(m.docPath)
		if err != nil {
			t.Errorf("%s: no doc file at %s", m.name, filepath.Base(m.docPath))
			continue
		}
		text := string(src)

		// A leading === is not a separator, because ParseDocFile splits on
		// "\n===\n". It becomes the name of the first entry and swallows the
		// procedure that follows, which is then unreachable from doc().
		if strings.HasPrefix(strings.TrimLeft(text, " \t"), "===\n") {
			t.Errorf("%s: doc file begins with ===, which consumes the first entry; "+
				"entries are separated by === rather than preceded by it", m.name)
		}

		// Some doc files use a marker vocabulary that ParseDocFile does not
		// read at all, so every entry in them is unreachable.
		for _, marker := range []string{"@module", "@function", "@signature", "@description"} {
			if strings.Contains(text, marker+" ") {
				t.Errorf("%s: doc file uses %s markers, which ParseDocFile does not read; "+
					"the ===-separated form is required", m.name, marker)
				break
			}
		}
	}
}

// TestLibDocCoversExports checks that a module doc file describes exactly the
// procedures the module exports.
func TestLibDocCoversExports(t *testing.T) {
	for _, m := range loadModules(t) {
		src, err := os.ReadFile(m.docPath)
		if err != nil {
			continue // reported by TestLibDocFilesWellFormed
		}
		// Strip @preamble lines before parsing.
		text := stripPreambleLines(string(src))
		docs, err := help.ParseDocFile(m.docPath, text)
		if err != nil {
			t.Errorf("%s: parsing doc: %v", m.name, err)
			continue
		}

		var missing []string
		for _, name := range m.exports {
			if _, ok := docs[name]; !ok {
				missing = append(missing, name)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("%s: exported but absent from doc: %s", m.name, strings.Join(missing, ", "))
		}

		exported := map[string]bool{}
		for _, name := range m.exports {
			exported[name] = true
		}
		var phantom []string
		for name := range docs {
			if name == "===" || !exported[name] {
				phantom = append(phantom, name)
			}
		}
		sort.Strings(phantom)
		if len(phantom) > 0 {
			t.Errorf("%s: doc describes procedures the module does not export: %s",
				m.name, strings.Join(phantom, ", "))
		}
	}
}

// TestLibModulesHaveHelp checks that every exporting module has a help file.
func TestLibModulesHaveHelp(t *testing.T) {
	for _, m := range loadModules(t) {
		if _, err := os.Stat(m.helpPath); err != nil {
			t.Errorf("%s: no help file at %s", m.name, filepath.Base(m.helpPath))
		}
	}
}

// TestShippedModuleDiscovery guards the discovery itself: these checks are
// worthless if the registry silently finds nothing, and they would be
// overreaching if they covered modules outside the distribution.
func TestShippedModuleDiscovery(t *testing.T) {
	paths := shippedModulePaths(t)

	if len(paths) < 10 {
		t.Errorf("expected the distribution to ship at least 10 modules, found %d", len(paths))
	}

	// A module a developer places outside the distribution roots is their own
	// business and must not be pulled into the documentation invariants.
	for name, path := range paths {
		rel := filepath.ToSlash(path)
		shipped := false
		for _, root := range substrate.DistributionModuleRoots() {
			if strings.Contains(rel, "/"+root+"/") {
				shipped = true
				break
			}
		}
		if !shipped {
			t.Errorf("%s at %s is outside the distribution roots %v",
				name, path, substrate.DistributionModuleRoots())
		}
	}
}
