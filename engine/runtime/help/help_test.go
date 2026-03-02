package help

import (
	"testing"
)

func TestParseHelpFile(t *testing.T) {
	source := `# Test help file

@func print
@template "print(val, ...)"
@help "Prints values to stdout."

@func first
@template "first(list)"
@help "Returns first element."
`

	entries, err := ParseHelpFile("test.help", source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if e, ok := entries["print"]; !ok {
		t.Error("missing print entry")
	} else {
		if e.Template != "print(val, ...)" {
			t.Errorf("print template = %q, want %q", e.Template, "print(val, ...)")
		}
		if e.Help != "Prints values to stdout." {
			t.Errorf("print help = %q, want %q", e.Help, "Prints values to stdout.")
		}
	}

	if e, ok := entries["first"]; !ok {
		t.Error("missing first entry")
	} else {
		if e.Template != "first(list)" {
			t.Errorf("first template = %q, want %q", e.Template, "first(list)")
		}
	}
}

func TestParseHelpFileMissingTemplate(t *testing.T) {
	source := `@func broken
@help "No template here."
`

	_, err := ParseHelpFile("test.help", source)
	if err == nil {
		t.Error("expected error for missing @template")
	}
}

func TestParseHelpFileMissingHelp(t *testing.T) {
	source := `@func broken
@template "broken()"
`

	_, err := ParseHelpFile("test.help", source)
	if err == nil {
		t.Error("expected error for missing @help")
	}
}

func TestParseHelpFileOrphanTemplate(t *testing.T) {
	source := `@template "orphan()"
@help "No func."
`

	_, err := ParseHelpFile("test.help", source)
	if err == nil {
		t.Error("expected error for @template without @func")
	}
}

func TestParseDocFile(t *testing.T) {
	source := `print
Prints values to stdout.

print(val, ...)

print("hello")
===
first
Returns first element of list.

first(list)

first([1, 2, 3])  # 1
`

	entries, err := ParseDocFile("test.doc", source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if e, ok := entries["print"]; !ok {
		t.Error("missing print entry")
	} else {
		if e.Doc == "" {
			t.Error("print doc is empty")
		}
	}

	if _, ok := entries["first"]; !ok {
		t.Error("missing first entry")
	}
}

func TestParseDocFileDuplicate(t *testing.T) {
	source := `print
First entry.
===
print
Duplicate entry.
`

	_, err := ParseDocFile("test.doc", source)
	if err == nil {
		t.Error("expected error for duplicate entry")
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	funcs := map[string]FuncEntry{
		"print": {Name: "print", Template: "print(val)", Help: "Prints."},
	}
	docs := map[string]DocEntry{
		"print": {Name: "print", Doc: "Full doc here."},
	}

	r.Merge(funcs, docs)

	if h := r.GetHelp("print"); h == nil {
		t.Error("GetHelp returned nil")
	} else if h.Help != "Prints." {
		t.Errorf("help = %q, want %q", h.Help, "Prints.")
	}

	if d := r.GetDoc("print"); d == nil {
		t.Error("GetDoc returned nil")
	} else if d.Doc != "Full doc here." {
		t.Errorf("doc = %q, want %q", d.Doc, "Full doc here.")
	}

	if r.GetHelp("missing") != nil {
		t.Error("GetHelp should return nil for missing")
	}

	names := r.ListFuncs()
	if len(names) != 1 || names[0] != "print" {
		t.Errorf("ListFuncs = %v, want [print]", names)
	}
}
