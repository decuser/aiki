package engine

// Observer receives events during lexing, parsing, evaluation, and formatting.
// All parameters are primitives to avoid coupling to value types.
type Observer interface {
	OnLex(token string, lexeme string, pos Position)
	OnParse(production string, depth int, pos Position)
	OnEval(node string, result string, scope int, pos Position)
	OnEffect(action string, target string, pos Position)
	OnFormat(method string, output string, node string, depth int)
}
