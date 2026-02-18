package syntax

import _ "embed"

//go:embed grammar.ebnf
var grammarSource string

var cachedGrammar *Grammar

func GetGrammar() *Grammar {
	if cachedGrammar == nil {
		g, err := Parse(grammarSource)
		if err != nil {
			panic(err)
		}
		cachedGrammar = g
	}
	return cachedGrammar
}

