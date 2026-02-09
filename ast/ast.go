package ast

import "aiki/token"

// Node is the interface all AST nodes implement.
type Node interface {
	TokenLiteral() string
}

// Statement nodes
type Statement interface {
	Node
	statementNode()
}

// Expression nodes
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// LetStatement: let x = expr
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (s *LetStatement) statementNode()       {}
func (s *LetStatement) TokenLiteral() string { return s.Token.Lexeme }

// ShapeStatement: let @point [x y]
type ShapeStatement struct {
	Token  token.Token
	Name   string
	Fields []string
	Embeds []string // embedded shape names
}

func (s *ShapeStatement) statementNode()       {}
func (s *ShapeStatement) TokenLiteral() string { return s.Token.Lexeme }

// AssignStatement: x = expr
type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (s *AssignStatement) statementNode()       {}
func (s *AssignStatement) TokenLiteral() string { return s.Token.Lexeme }

// ReturnStatement: return expr
type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (s *ReturnStatement) statementNode()       {}
func (s *ReturnStatement) TokenLiteral() string { return s.Token.Lexeme }

// ExpressionStatement wraps an expression as a statement.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (s *ExpressionStatement) statementNode()       {}
func (s *ExpressionStatement) TokenLiteral() string { return s.Token.Lexeme }

// BlockStatement: { statements }
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (s *BlockStatement) statementNode()       {}
func (s *BlockStatement) TokenLiteral() string { return s.Token.Lexeme }

// IfStatement: if cond { } else { }
type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (s *IfStatement) statementNode()       {}
func (s *IfStatement) TokenLiteral() string { return s.Token.Lexeme }

// WhileStatement: while cond { }
type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (s *WhileStatement) statementNode()       {}
func (s *WhileStatement) TokenLiteral() string { return s.Token.Lexeme }

// MatchStatement: match expr { arms }
type MatchStatement struct {
	Token token.Token
	Value Expression
	Arms  []*MatchArm
}

func (s *MatchStatement) statementNode()       {}
func (s *MatchStatement) TokenLiteral() string { return s.Token.Lexeme }

// MatchArm: pattern { body }
type MatchArm struct {
	Pattern Pattern
	Body    *BlockStatement
}

// Pattern for match arms
type Pattern interface {
	Node
	patternNode()
}

// WildcardPattern: _
type WildcardPattern struct {
	Token token.Token
}

func (p *WildcardPattern) patternNode()         {}
func (p *WildcardPattern) TokenLiteral() string { return p.Token.Lexeme }

// NamePattern: binds to a name
type NamePattern struct {
	Token token.Token
	Name  string
}

func (p *NamePattern) patternNode()         {}
func (p *NamePattern) TokenLiteral() string { return p.Token.Lexeme }

// LiteralPattern: matches a literal value
type LiteralPattern struct {
	Token token.Token
	Value Expression
}

func (p *LiteralPattern) patternNode()         {}
func (p *LiteralPattern) TokenLiteral() string { return p.Token.Lexeme }

// ListPattern: [p1 p2 p3]
type ListPattern struct {
	Token    token.Token
	Elements []Pattern
}

func (p *ListPattern) patternNode()         {}
func (p *ListPattern) TokenLiteral() string { return p.Token.Lexeme }

// ShapedListPattern: [@name p1 p2]
type ShapedListPattern struct {
	Token    token.Token
	Shape    string
	Elements []Pattern
}

func (p *ShapedListPattern) patternNode()         {}
func (p *ShapedListPattern) TokenLiteral() string { return p.Token.Lexeme }

// ExportStatement: export [name1 name2]
type ExportStatement struct {
	Token token.Token
	Names []string
}

func (s *ExportStatement) statementNode()       {}
func (s *ExportStatement) TokenLiteral() string { return s.Token.Lexeme }

// ImportStatement: from mod use [name1 name2]
type ImportStatement struct {
	Token  token.Token
	Module string
	Names  []string
}

func (s *ImportStatement) statementNode()       {}
func (s *ImportStatement) TokenLiteral() string { return s.Token.Lexeme }

// Expressions

// Identifier: x, foo_bar
type Identifier struct {
	Token token.Token
	Value string
}

func (e *Identifier) expressionNode()      {}
func (e *Identifier) TokenLiteral() string { return e.Token.Lexeme }

// NumberLiteral: 42, 3.14, 1/3
type NumberLiteral struct {
	Token token.Token
	Value string // raw string, parsed later to rational
}

func (e *NumberLiteral) expressionNode()      {}
func (e *NumberLiteral) TokenLiteral() string { return e.Token.Lexeme }

// BooleanLiteral: true, false
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (e *BooleanLiteral) expressionNode()      {}
func (e *BooleanLiteral) TokenLiteral() string { return e.Token.Lexeme }

// StringLiteral: "hello"
type StringLiteral struct {
	Token token.Token
	Value string
}

func (e *StringLiteral) expressionNode()      {}
func (e *StringLiteral) TokenLiteral() string { return e.Token.Lexeme }

// RuneLiteral: 'a'
type RuneLiteral struct {
	Token token.Token
	Value rune
}

func (e *RuneLiteral) expressionNode()      {}
func (e *RuneLiteral) TokenLiteral() string { return e.Token.Lexeme }

// SymbolLiteral: :active
type SymbolLiteral struct {
	Token token.Token
	Value string
}

func (e *SymbolLiteral) expressionNode()      {}
func (e *SymbolLiteral) TokenLiteral() string { return e.Token.Lexeme }

// ListLiteral: [1 2 3]
type ListLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (e *ListLiteral) expressionNode()      {}
func (e *ListLiteral) TokenLiteral() string { return e.Token.Lexeme }

// ShapedListLiteral: [@point 10 20]
type ShapedListLiteral struct {
	Token    token.Token
	Shape    string
	Elements []Expression
}

func (e *ShapedListLiteral) expressionNode()      {}
func (e *ShapedListLiteral) TokenLiteral() string { return e.Token.Lexeme }

// FunctionLiteral: (x y) { return x + y }
type FunctionLiteral struct {
	Token      token.Token
	Parameters []string
	Body       *BlockStatement
}

func (e *FunctionLiteral) expressionNode()      {}
func (e *FunctionLiteral) TokenLiteral() string { return e.Token.Lexeme }

// CallExpression: f(x, y)
type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (e *CallExpression) expressionNode()      {}
func (e *CallExpression) TokenLiteral() string { return e.Token.Lexeme }

// AccessExpression: list.0, point.x
type AccessExpression struct {
	Token token.Token
	Left  Expression
	Key   string
}

func (e *AccessExpression) expressionNode()      {}
func (e *AccessExpression) TokenLiteral() string { return e.Token.Lexeme }

// InfixExpression: a + b
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (e *InfixExpression) expressionNode()      {}
func (e *InfixExpression) TokenLiteral() string { return e.Token.Lexeme }

// PrefixExpression: not x
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (e *PrefixExpression) expressionNode()      {}
func (e *PrefixExpression) TokenLiteral() string { return e.Token.Lexeme }

// PipeExpression: x |> f() |> g()
type PipeExpression struct {
	Token token.Token
	Left  Expression
	Right *CallExpression
}

func (e *PipeExpression) expressionNode()      {}
func (e *PipeExpression) TokenLiteral() string { return e.Token.Lexeme }
