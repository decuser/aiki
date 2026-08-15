package syntax

import "aiki/engine/syntax/grammar"

// NormalizeTokens applies grammar-declared skip and physical-newline policy to
// a raw lexer stream. It is the observable normalized token surface used by
// parser construction and cross-implementation conformance.
func NormalizeTokens(g *grammar.Grammar, tokens []Token) []Token {
	filtered, _ := normalizeTokens(g, tokens)
	return filtered
}

// normalizeTokens also returns the indexes of semicolons synthesized from
// physical newlines. Parser diagnostics retain that provenance internally.
func normalizeTokens(g *grammar.Grammar, tokens []Token) ([]Token, map[int]bool) {
	rule := g.Newline
	afterToken := make(map[string]struct{}, len(rule.AfterToken))
	for _, name := range rule.AfterToken {
		afterToken[name] = struct{}{}
	}
	afterLexeme := make(map[string]struct{}, len(rule.AfterLexeme))
	for _, lexeme := range rule.AfterLexeme {
		afterLexeme[lexeme] = struct{}{}
	}
	suppressOpen := make(map[string]string, len(rule.SuppressIn))
	suppressClose := make(map[string]struct{}, len(rule.SuppressIn))
	for _, pair := range rule.SuppressIn {
		suppressOpen[pair[0]] = pair[1]
		suppressClose[pair[1]] = struct{}{}
	}
	endsStatement := func(tok Token) bool {
		if _, ok := afterToken[tok.Type]; ok {
			return true
		}
		_, ok := afterLexeme[tok.Lexeme]
		return ok
	}

	filtered := make([]Token, 0, len(tokens))
	syntheticTerminators := make(map[int]bool)
	var suppressStack []string
	for _, tok := range tokens {
		if def, ok := g.GetToken(tok.Type); ok && def.Skip {
			continue
		}

		// Suppression is delimiter-aware rather than an aggregate depth. An
		// unmatched or mismatched closer cannot drive state negative or cancel a
		// later legitimate opener; the grammar parser reports malformed
		// delimiters downstream.
		if len(suppressStack) > 0 && tok.Lexeme == suppressStack[len(suppressStack)-1] {
			suppressStack = suppressStack[:len(suppressStack)-1]
		} else if closer, ok := suppressOpen[tok.Lexeme]; ok {
			suppressStack = append(suppressStack, closer)
		} else if _, ok := suppressClose[tok.Lexeme]; ok {
			// Unmatched/mismatched closer: ignore for normalization state.
		}

		if tok.Type == rule.Token {
			if len(suppressStack) == 0 && len(filtered) > 0 && endsStatement(filtered[len(filtered)-1]) {
				syntheticTerminators[len(filtered)] = true
				filtered = append(filtered, Token{
					Type:   "DELIMITER",
					Lexeme: ";",
					Pos:    tok.Pos,
				})
			}
			continue
		}
		filtered = append(filtered, tok)
	}
	return filtered, syntheticTerminators
}
