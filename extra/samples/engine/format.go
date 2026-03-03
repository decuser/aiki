package engine

import (
	"fmt"
	"strings"
)

// FormatWithCaret formats an error location with source line and caret.
//
// Output:
//
//	file.ai:4:10: message
//	    let x = 1 / 0
//	             ^
func FormatWithCaret(pos Position, source string, message string) string {
	var sb strings.Builder

	// Header: file:line:col: message
	if pos.File != "" {
		sb.WriteString(fmt.Sprintf("%s:%d:%d: %s", pos.File, pos.Line, pos.Col, message))
	} else {
		sb.WriteString(fmt.Sprintf("%d:%d: %s", pos.Line, pos.Col, message))
	}

	// Source line and caret
	if source != "" {
		sb.WriteString("\n    ")
		sb.WriteString(strings.TrimRight(source, "\r\n"))
		sb.WriteString("\n    ")
		// Position caret under the column
		if pos.Col > 0 {
			sb.WriteString(strings.Repeat(" ", pos.Col-1))
		}
		sb.WriteString("^")
	}

	return sb.String()
}

// GetSourceLine extracts a specific line from source text.
// Line is 1-indexed. Returns empty string if out of range.
func GetSourceLine(source string, line int) string {
	if line < 1 {
		return ""
	}
	lines := strings.Split(source, "\n")
	if line > len(lines) {
		return ""
	}
	return lines[line-1]
}

// FormatCaret returns a caret string positioned at the given column.
func FormatCaret(col int) string {
	if col <= 0 {
		return "^"
	}
	return strings.Repeat(" ", col-1) + "^"
}
