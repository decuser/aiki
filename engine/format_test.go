package engine

import (
	"strings"
	"testing"
)

func TestFormatWithCaret(t *testing.T) {
	pos := Position{File: "test.ai", Line: 4, Col: 10}
	source := "let x = 1 / 0"
	msg := "division by zero"

	result := FormatWithCaret(pos, source, msg)

	// Check header
	if !strings.Contains(result, "test.ai:4:10: division by zero") {
		t.Errorf("expected header, got:\n%s", result)
	}

	// Check source line present
	if !strings.Contains(result, "let x = 1 / 0") {
		t.Errorf("expected source line, got:\n%s", result)
	}

	// Check caret position (9 spaces + ^)
	if !strings.Contains(result, "         ^") {
		t.Errorf("expected caret at col 10, got:\n%s", result)
	}
}

func TestFormatWithCaretNoFile(t *testing.T) {
	pos := Position{Line: 1, Col: 5}
	source := "1 + + 2"
	msg := "unexpected '+'"

	result := FormatWithCaret(pos, source, msg)

	if !strings.HasPrefix(result, "1:5:") {
		t.Errorf("expected line:col format, got:\n%s", result)
	}
}

func TestFormatWithCaretNoSource(t *testing.T) {
	pos := Position{File: "test.ai", Line: 1, Col: 1}
	msg := "something went wrong"

	result := FormatWithCaret(pos, "", msg)

	if strings.Contains(result, "\n") {
		t.Errorf("expected single line without source, got:\n%s", result)
	}
	if result != "test.ai:1:1: something went wrong" {
		t.Errorf("unexpected output: %s", result)
	}
}

func TestFormatWithCaretCol1(t *testing.T) {
	pos := Position{File: "test.ai", Line: 1, Col: 1}
	source := "let x = 1"
	msg := "error at start"

	result := FormatWithCaret(pos, source, msg)

	// Caret should be at start (no leading spaces)
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	caretLine := lines[2]
	if caretLine != "    ^" {
		t.Errorf("expected caret at col 1, got: '%s'", caretLine)
	}
}

func TestGetSourceLine(t *testing.T) {
	source := "line one\nline two\nline three"

	tests := []struct {
		line     int
		expected string
	}{
		{1, "line one"},
		{2, "line two"},
		{3, "line three"},
		{0, ""},
		{4, ""},
		{-1, ""},
	}

	for _, tt := range tests {
		result := GetSourceLine(source, tt.line)
		if result != tt.expected {
			t.Errorf("GetSourceLine(%d) = %q, want %q", tt.line, result, tt.expected)
		}
	}
}
