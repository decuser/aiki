// Package evaluator implements the Aiki evaluation engine.
// walker.go provides AST traversal helpers for semantic walking.
package evaluator

import "aiki/engine/syntax"

// findChild returns the first child matching the given type, or nil.
func findChild(node *syntax.Node, typ string) *syntax.Node {
	for _, child := range node.Children {
		if child.Type == typ {
			return child
		}
	}
	return nil
}

// findChildValue returns the value of the first child matching the given type.
func findChildValue(node *syntax.Node, typ string) (string, bool) {
	if child := findChild(node, typ); child != nil {
		return child.Value, true
	}
	return "", false
}

// findChildren returns all children matching the given type.
func findChildren(node *syntax.Node, typ string) []*syntax.Node {
	var result []*syntax.Node
	for _, child := range node.Children {
		if child.Type == typ {
			result = append(result, child)
		}
	}
	return result
}

// findAfterTerminal returns the node immediately following a terminal with the given value.
func findAfterTerminal(node *syntax.Node, terminal string) *syntax.Node {
	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == terminal && i+1 < len(node.Children) {
			return node.Children[i+1]
		}
	}
	return nil
}

// hasTerminal returns true if the node contains a terminal with the given value.
func hasTerminal(node *syntax.Node, terminal string) bool {
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == terminal {
			return true
		}
	}
	return false
}

// nonTerminalChildren returns all children that are not TERMINAL nodes.
func nonTerminalChildren(node *syntax.Node) []*syntax.Node {
	var result []*syntax.Node
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			result = append(result, child)
		}
	}
	return result
}

// firstNonTerminal returns the first non-TERMINAL child, or nil.
func firstNonTerminal(node *syntax.Node) *syntax.Node {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			return child
		}
	}
	return nil
}

// isTerminal checks if a node is a TERMINAL node.
func isTerminal(node *syntax.Node) bool {
	return node.Type == "TERMINAL"
}

// terminalValue returns the value if this is a TERMINAL, empty string otherwise.
func terminalValue(node *syntax.Node) string {
	if node.Type == "TERMINAL" {
		return node.Value
	}
	return ""
}

// walkChildren iterates over non-terminal children, calling fn for each.
// Returns early if fn returns an error.
func walkChildren(node *syntax.Node, fn func(*syntax.Node) error) error {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			if err := fn(child); err != nil {
				return err
			}
		}
	}
	return nil
}

// nodePosition returns the line number from a node, traversing into children if needed.
func nodePosition(node *syntax.Node) int {
	if node.Pos.Line > 0 {
		return node.Pos.Line
	}
	for _, child := range node.Children {
		if line := nodePosition(child); line > 0 {
			return line
		}
	}
	return 0
}
