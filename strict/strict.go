package strict

import (
	_ "embed"
	"fmt"

	"aiki/lang/ast"
	"aiki/lang/eval"
	"aiki/lang/parser"
	"aiki/lang/value"
)

//go:embed strict.ai
var strictSource string

func LoadStrict(env *value.Env) error {
	result := eval.Run(strictSource, env)
	if e, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", e.Message)
	}
	env.SnapshotStrict()
	return nil
}

func Exports() []string {
	p := parser.New(strictSource)
	program := p.Parse()
	for _, stmt := range program.Statements {
		if exp, ok := stmt.(*ast.ExportStatement); ok {
			return exp.Names
		}
	}
	return nil
}
