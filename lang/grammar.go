package lang

import (
    _ "embed"
    "aiki/internal/ebnf"
)

//go:embed grammar.ebnf
var grammarSource string

var grammar *ebnf.Grammar

func Grammar() *ebnf.Grammar {
    if grammar == nil {
        g, err := ebnf.Parse(grammarSource)
        if err != nil {
            panic(err)
        }
        grammar = g
    }
    return grammar
}
