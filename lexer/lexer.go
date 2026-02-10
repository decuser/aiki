package lexer

import (
	"unicode"
	"unicode/utf8"

	"aiki/token"
)

const eof rune = -1

type state int

const (
	stateStart state = iota
	stateNumber
	stateString
	stateRune
	stateSymbol
	stateShape
	stateName
	stateComment
	stateDone
)

// Comment represents a source comment with its line number.
type Comment struct {
	Line  int
	Text  string
	IsEOL bool // true if comment follows code on same line
}

type Lexer struct {
	input       string
	start       int
	pos         int
	width       int
	line        int
	col         int
	state       state
	tokens      []token.Token
	Comments    []Comment // collected comments
	lastTokLine int       // line of last emitted token
}

func New(input string) *Lexer {
	return &Lexer{
		input: input,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) Tokenize() []token.Token {
	l.state = stateStart
	for l.state != stateDone {
		switch l.state {
		case stateStart:
			l.state = l.lexStart()
		case stateNumber:
			l.state = l.lexNumber()
		case stateString:
			l.state = l.lexString()
		case stateRune:
			l.state = l.lexRune()
		case stateSymbol:
			l.state = l.lexSymbol()
		case stateShape:
			l.state = l.lexShape()
		case stateName:
			l.state = l.lexName()
		case stateComment:
			l.state = l.lexComment()
		}
	}
	l.emit(token.EOF)
	return l.tokens
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) backup() {
	l.pos -= l.width
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	if r == '\n' {
		l.line--
	} else {
		l.col--
	}
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) emit(t token.Type) {
	l.tokens = append(l.tokens, token.Token{
		Type:   t,
		Lexeme: l.input[l.start:l.pos],
		Start:  l.start,
		End:    l.pos,
		Line:   l.line,
		Col:    l.col,
	})
	l.lastTokLine = l.line
	l.start = l.pos
}

func (l *Lexer) lexStart() state {
	r := l.next()
	if r == eof {
		return stateDone
	}

	// Skip whitespace (except newlines for now)
	if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
		l.start = l.pos
		return stateStart
	}

	// Comment
	if r == '#' {
		return stateComment
	}

	// String
	if r == '"' {
		return stateString
	}

	// Rune
	if r == '\'' {
		return stateRune
	}

	// Symbol
	if r == ':' {
		return stateSymbol
	}

	// Shape
	if r == '@' {
		return stateShape
	}

	// Number
	if isDigit(r) {
		return stateNumber
	}

	// Name or keyword
	if isLetter(r) || r == '_' {
		return stateName
	}

	// Operators and delimiters
	switch r {
	case '+':
		l.emit(token.Plus)
	case '-':
		l.emit(token.Minus)
	case '*':
		l.emit(token.Star)
	case '/':
		l.emit(token.Slash)
	case '%':
		l.emit(token.Percent)
	case '=':
		if l.peek() == '=' {
			l.next()
			l.emit(token.Eq)
		} else {
			l.emit(token.Assign)
		}
	case '!':
		if l.peek() == '=' {
			l.next()
			l.emit(token.NotEq)
		} else {
			l.emit(token.Illegal)
		}
	case '<':
		if l.peek() == '=' {
			l.next()
			l.emit(token.LtEq)
		} else {
			l.emit(token.Lt)
		}
	case '>':
		if l.peek() == '=' {
			l.next()
			l.emit(token.GtEq)
		} else {
			l.emit(token.Gt)
		}
	case '|':
		if l.peek() == '>' {
			l.next()
			l.emit(token.Pipe)
		} else {
			l.emit(token.Illegal)
		}
	case '.':
		l.emit(token.Dot)
	case '(':
		l.emit(token.LParen)
	case ')':
		l.emit(token.RParen)
	case '[':
		l.emit(token.LBracket)
	case ']':
		l.emit(token.RBracket)
	case '{':
		l.emit(token.LBrace)
	case '}':
		l.emit(token.RBrace)
	default:
		l.emit(token.Illegal)
	}

	return stateStart
}

func (l *Lexer) lexNumber() state {
	// Already consumed first digit
	for isDigit(l.peek()) {
		l.next()
	}

	// Check for decimal: 3.14
	if l.peek() == '.' {
		l.next()
		if isDigit(l.peek()) {
			for isDigit(l.peek()) {
				l.next()
			}
		} else {
			// Trailing dot, backup
			l.backup()
		}
	}

	l.emit(token.Number)
	return stateStart
}

func (l *Lexer) lexString() state {
	// Opening quote already consumed
	for {
		r := l.next()
		if r == eof {
			l.emit(token.Illegal)
			return stateDone
		}
		if r == '\\' {
			l.next() // skip escaped char
			continue
		}
		if r == '"' {
			break
		}
	}
	l.emit(token.String)
	return stateStart
}

func (l *Lexer) lexRune() state {
	// Opening quote already consumed
	r := l.next()
	if r == '\\' {
		l.next() // escaped char
	}
	r = l.next()
	if r != '\'' {
		l.emit(token.Illegal)
		return stateStart
	}
	l.emit(token.Rune)
	return stateStart
}

func (l *Lexer) lexSymbol() state {
	// Colon already consumed
	if !isLetter(l.peek()) && l.peek() != '_' {
		l.emit(token.Illegal)
		return stateStart
	}
	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' {
		l.next()
	}
	l.emit(token.Symbol)
	return stateStart
}

func (l *Lexer) lexShape() state {
	// @ already consumed
	if !isLetter(l.peek()) && l.peek() != '_' {
		l.emit(token.Illegal)
		return stateStart
	}
	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' {
		l.next()
	}
	l.emit(token.Shape)
	return stateStart
}

func (l *Lexer) lexName() state {
	// First char already consumed
	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' {
		l.next()
	}
	lexeme := l.input[l.start:l.pos]
	t := token.LookupName(lexeme)
	l.emit(t)
	return stateStart
}

func (l *Lexer) lexComment() state {
	// # already consumed
	commentLine := l.line
	isEOL := l.lastTokLine == commentLine // comment on same line as previous token
	start := l.pos
	for {
		r := l.next()
		if r == eof || r == '\n' {
			break
		}
	}
	// Save comment (without the newline)
	text := l.input[start : l.pos-l.width]
	if l.width == 0 {
		// EOF case
		text = l.input[start:l.pos]
	}
	l.Comments = append(l.Comments, Comment{
		Line:  commentLine,
		Text:  "#" + text,
		IsEOL: isEOL,
	})
	l.start = l.pos
	return stateStart
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}
