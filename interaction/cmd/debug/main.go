package main

import (
	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/syntax"
	"aiki/engine/syntax/definition"
	"flag"
	"fmt"
	"os"
)

func main() {
	boundary := flag.String("boundary", "all", "filter output by boundary")
	trace := flag.Bool("trace", false, "enable trace observation")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: debug <filename>")
		return
	}

	path := flag.Arg(0)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return
	}

	grammar := definition.New()
	var obs engine.Observer
	if *trace {
		obs = TraceObserver{}
		grammar.SetObserver(obs)
	}

	lexer := syntax.NewLexer(path, string(data), grammar)
	parser, err := syntax.NewParser(lexer, grammar)
	if err != nil {
		fmt.Printf("parser init failed: %v\n", err)
		return
	}
	tree, err := parser.Parse()
	if err != nil {
		fmt.Printf("parse error: %v\n", err)
		return
	}

	if *boundary == "all" || *boundary == "eval" {
		fmt.Println("\n==== Entering Evaluator ====")
		rt := substrate.NewGoRuntime()
		global := evaluator.NewScope(nil)
		global.SetFile(path)
		global.SetSource(string(data))
		prelude.Load(global, rt)

		ev := evaluator.New(rt, obs)
		ev.SetGrammar(grammar)
		result, err := ev.Eval(tree, global)
		if err != nil {
			fmt.Printf("eval error: %v\n", err)
		} else {
			fmt.Printf("=> %s\n", result.Inspect())
		}
	}
}
