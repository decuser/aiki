package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/syntax/grammar"
)

// FFIUsage reports statically reachable FFI module realizations from an Aiki
// source file. It follows literal import()/use() calls through named modules and
// explicit relative-path imports. Dynamic module names are intentionally not
// guessed.
func FFIUsage(g *grammar.Grammar, entry string, roots []string) ([]string, error) {
	registry := NewModuleRegistry(roots)
	if err := registry.Scan(g); err != nil {
		return nil, err
	}

	seenFiles := make(map[string]bool)
	found := make(map[string]bool)
	var visitFile func(string) error
	visitFile = func(path string) error {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
		if real, err := filepath.EvalSymlinks(path); err == nil {
			path = real
		}
		path = filepath.Clean(path)
		if seenFiles[path] {
			return nil
		}
		seenFiles[path] = true

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := AnalyzeSource(g, path, string(data))
		if err != nil {
			return err
		}
		for _, name := range info.Imports {
			if IsPathImport(name) {
				local := name
				if !strings.HasSuffix(local, ".ai") {
					local += ".ai"
				}
				local = filepath.Join(filepath.Dir(path), local)
				if _, err := os.Stat(local); err != nil {
					return fmt.Errorf("%s imports missing path %q", path, name)
				}
				if err := visitFile(local); err != nil {
					return err
				}
				continue
			}

			modulePath, canonical, ok := registry.Resolve(name)
			if !ok {
				return fmt.Errorf("%s imports unknown package %q", path, name)
			}
			policy, declared := StdlibModulePolicyFor(canonical)
			if (declared && policy.Realization == RealizationFFI) || strings.HasSuffix(canonical, "/ffi") {
				found[canonical] = true
			}
			if err := visitFile(modulePath); err != nil {
				return err
			}
		}
		return nil
	}

	if err := visitFile(entry); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(found))
	for name := range found {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}
