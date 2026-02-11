package token

type Type int

const (
	Illegal Type = iota
	EOF

	// Literals
	Number // 42, 3.14, 1/3
	String // "hello"
	Rune   // 'a'
	Symbol // :active
	Shape  // @point
	Name   // x, square

	// Keywords
	Let
	If
	Else
	While
	Match
	Return
	Export
	From
	Use
	True
	False
	And
	Or
	Not

	// Operators
	Plus    // +
	Minus   // -
	Star    // *
	Slash   // /
	Percent // %
	Eq      // ==
	NotEq   // !=
	Lt      // <
	Gt      // >
	LtEq    // <=
	GtEq    // >=
	Assign  // =
	Pipe    // |>
	Dot     // .

	// Delimiters
	Comma    // ,
	LParen   // (
	RParen   // )
	LBracket // [
	RBracket // ]
	LBrace   // {
	RBrace   // }

	// Special
	Newline
	Comment
)

var keywords = map[string]Type{
	"let":    Let,
	"if":     If,
	"else":   Else,
	"while":  While,
	"match":  Match,
	"return": Return,
	"export": Export,
	"from":   From,
	"use":    Use,
	"true":   True,
	"false":  False,
	"and":    And,
	"or":     Or,
	"not":    Not,
}

func LookupName(name string) Type {
	if t, ok := keywords[name]; ok {
		return t
	}
	return Name
}

type Token struct {
	Type   Type
	Lexeme string
	Start  int
	End    int
	Line   int
	Col    int
}

func (t Token) String() string {
	return t.Lexeme
}

var typeNames = map[Type]string{
	Illegal:  "Illegal",
	EOF:      "EOF",
	Number:   "Number",
	String:   "String",
	Rune:     "Rune",
	Symbol:   "Symbol",
	Shape:    "Shape",
	Name:     "Name",
	Let:      "Let",
	If:       "If",
	Else:     "Else",
	While:    "While",
	Match:    "Match",
	Return:   "Return",
	Export:   "Export",
	From:     "From",
	Use:      "Use",
	True:     "True",
	False:    "False",
	And:      "And",
	Or:       "Or",
	Not:      "Not",
	Plus:     "Plus",
	Minus:    "Minus",
	Star:     "Star",
	Slash:    "Slash",
	Percent:  "Percent",
	Eq:       "Eq",
	NotEq:    "NotEq",
	Lt:       "Lt",
	Gt:       "Gt",
	LtEq:     "LtEq",
	GtEq:     "GtEq",
	Assign:   "Assign",
	Pipe:     "Pipe",
	Dot:      "Dot",
	Comma:    "Comma",
	LParen:   "LParen",
	RParen:   "RParen",
	LBracket: "LBracket",
	RBracket: "RBracket",
	LBrace:   "LBrace",
	RBrace:   "RBrace",
	Newline:  "Newline",
	Comment:  "Comment",
}

func (t Type) String() string {
	if name, ok := typeNames[t]; ok {
		return name
	}
	return "Unknown"
}
