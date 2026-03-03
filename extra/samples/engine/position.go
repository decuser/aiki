// Package engine provides core types shared across engine components.
package engine

// Position represents a source location.
type Position struct {
	File string
	Line int
	Col  int
}
