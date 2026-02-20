package main

import (
	"fmt"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/definition"
	"aiki/engine"
)

// silentObserver implements engine.Observer with no-ops
type silentObserver struct{}

func (s silentObserver) OnLex(token string, lexeme string, pos engine.Position)            {}
func (s silentObserver) OnParse(production string, depth int, pos engine.Position)         {}
func (s silentObserver) OnReady(substrate string, scope int)                               {}
func (s silentObserver) OnEval(op string, val value.Value, scope int, pos engine.Position) {}
func (s silentObserver) OnEffect(action string, substrate string, pos engine.Position)     {}

func run(src string) {
	g := definition.New()
	g.SetObserver(silentObserver{})
	rt := substrate.NewGoRuntime()

	lexer := syntax.NewLexer("test", src, g)
	parser, err := syntax.NewParser(lexer, g)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	ast, err := parser.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	e := evaluator.New(rt, silentObserver{})
	e.SetGrammar(g)
	scope := evaluator.NewScope(nil)

	result, err := e.Eval(ast, scope)
	if err != nil {
		fmt.Printf("Eval error: %v\n", err)
	} else {
		fmt.Printf("=> %s\n", result.Inspect())
	}
}

func main() {
	fmt.Println("=== Basic arithmetic ===")
	run("1 + 2 * 3")

	fmt.Println("\n=== Unary minus ===")
	run("-5")

	fmt.Println("\n=== Booleans ===")
	run("true")
	run("1 < 2")

	fmt.Println("\n=== Let and variables ===")
	run("let x = 10\nx")

	fmt.Println("\n=== Function call ===")
	run(`let double = (x) { x * 2 }
double(21)`)

	fmt.Println("\n=== Recursion (factorial) ===")
	run(`let factorial = (n) {
    if n <= 1 { return 1 }
    return n * factorial(n - 1)
}
factorial(5)`)

	fmt.Println("\n=== Closure ===")
	run(`let makeAdder = (x) {
    return (y) { return x + y }
}
let addTen = makeAdder(10)
addTen(5)`)

	fmt.Println("\n=== While loop ===")
	run(`let x = 0
while x < 5 {
    x = x + 1
}
x`)
}
