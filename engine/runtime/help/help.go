// Package help provides function documentation parsing and lookup.
package help

import (
	"fmt"
	"strings"
)

// FuncEntry holds help info for a function.
type FuncEntry struct {
	Name     string
	Template string // @template - signature
	Help     string // @help - one-liner
}

// DocEntry holds full documentation for a function.
type DocEntry struct {
	Name string
	Doc  string // full doc text
}

// Registry holds help and doc entries for lookup.
type Registry struct {
	Funcs map[string]FuncEntry
	Docs  map[string]DocEntry
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		Funcs: make(map[string]FuncEntry),
		Docs:  make(map[string]DocEntry),
	}
}

// ParseHelpFile parses a .help file with @func, @template, @help entries.
func ParseHelpFile(file, source string) (map[string]FuncEntry, error) {
	entries := make(map[string]FuncEntry)
	
	lines := strings.Split(source, "\n")
	var current *FuncEntry
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "@func ") {
			// Save previous entry
			if current != nil {
				if current.Template == "" || current.Help == "" {
					return nil, fmt.Errorf("%s:%d: @func %s missing @template or @help", file, i+1, current.Name)
				}
				entries[current.Name] = *current
			}
			
			name := strings.TrimSpace(strings.TrimPrefix(line, "@func"))
			if name == "" {
				return nil, fmt.Errorf("%s:%d: @func missing name", file, i+1)
			}
			current = &FuncEntry{Name: name}
			
		} else if strings.HasPrefix(line, "@template ") {
			if current == nil {
				return nil, fmt.Errorf("%s:%d: @template without @func", file, i+1)
			}
			current.Template = parseQuoted(strings.TrimPrefix(line, "@template"))
			
		} else if strings.HasPrefix(line, "@help ") {
			if current == nil {
				return nil, fmt.Errorf("%s:%d: @help without @func", file, i+1)
			}
			current.Help = parseQuoted(strings.TrimPrefix(line, "@help"))
			
		} else {
			return nil, fmt.Errorf("%s:%d: unexpected line: %s", file, i+1, line)
		}
	}
	
	// Save last entry
	if current != nil {
		if current.Template == "" || current.Help == "" {
			return nil, fmt.Errorf("%s: @func %s missing @template or @help", file, current.Name)
		}
		entries[current.Name] = *current
	}
	
	return entries, nil
}

// ParseDocFile parses a .doc file with === separated entries.
func ParseDocFile(file, source string) (map[string]DocEntry, error) {
	entries := make(map[string]DocEntry)
	parts := strings.Split(source, "\n===\n")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		lines := strings.SplitN(part, "\n", 2)
		name := strings.TrimSpace(lines[0])
		if name == "" {
			continue
		}
		
		var doc string
		if len(lines) > 1 {
			doc = strings.TrimSpace(lines[1])
		}
		
		if _, exists := entries[name]; exists {
			return nil, fmt.Errorf("%s: duplicate entry for '%s'", file, name)
		}
		entries[name] = DocEntry{Name: name, Doc: doc}
	}
	
	return entries, nil
}

// parseQuoted removes surrounding quotes if present.
func parseQuoted(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// Merge adds entries from help and doc maps into the registry.
func (r *Registry) Merge(funcs map[string]FuncEntry, docs map[string]DocEntry) {
	for name, entry := range funcs {
		r.Funcs[name] = entry
	}
	for name, entry := range docs {
		r.Docs[name] = entry
	}
}

// GetHelp returns the help entry for a name, or nil if not found.
func (r *Registry) GetHelp(name string) *FuncEntry {
	if entry, ok := r.Funcs[name]; ok {
		return &entry
	}
	return nil
}

// GetDoc returns the doc entry for a name, or nil if not found.
func (r *Registry) GetDoc(name string) *DocEntry {
	if entry, ok := r.Docs[name]; ok {
		return &entry
	}
	return nil
}

// ListFuncs returns all function names in the registry.
func (r *Registry) ListFuncs() []string {
	names := make([]string, 0, len(r.Funcs))
	for name := range r.Funcs {
		names = append(names, name)
	}
	return names
}
