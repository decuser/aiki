// Package syntax provides lexer, parser, and AST types for Aiki.
package syntax

import "aiki/engine"

// Node represents an AST node.
type Node struct {
	Type     string
	Value    string
	Children []*Node
	Pos      engine.Position
}

// IsTerminal returns true if this is a terminal node.
func (n *Node) IsTerminal() bool {
	return n.Type == "TERMINAL"
}
