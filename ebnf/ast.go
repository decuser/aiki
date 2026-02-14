package ebnf

// Token represents a lexed token from source
type Token struct {
	Type   string
	Lexeme string
	Line   int
	Column int
}

// Node represents a generic AST node
type Node struct {
	Type     string  // production name or token type
	Value    string  // lexeme for terminals
	Children []*Node // for non-terminals
	Line     int
	Column   int
}

// Helper methods for Node

// Child returns the nth child or nil
func (n *Node) Child(i int) *Node {
	if i < 0 || i >= len(n.Children) {
		return nil
	}
	return n.Children[i]
}

// ChildByType returns the first child with matching type
func (n *Node) ChildByType(typ string) *Node {
	for _, c := range n.Children {
		if c.Type == typ {
			return c
		}
	}
	return nil
}

// ChildrenByType returns all children with matching type
func (n *Node) ChildrenByType(typ string) []*Node {
	var result []*Node
	for _, c := range n.Children {
		if c.Type == typ {
			result = append(result, c)
		}
	}
	return result
}

// IsTerminal returns true if this is a leaf node
func (n *Node) IsTerminal() bool {
	return len(n.Children) == 0
}

// String returns a simple representation for debugging
func (n *Node) String() string {
	if n.IsTerminal() {
		return n.Type + ":" + n.Value
	}
	return n.Type
}
