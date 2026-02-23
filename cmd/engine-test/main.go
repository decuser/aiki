//go:build ignore

package main

import (
	"fmt"
	"os"

	"aiki/engine/runner"
)

func main() {
	if len(os.Args) < 2 {
		// REPL mode - just eval expressions
		fmt.Println("aiki-engine test runner")
		fmt.Println("Usage: go run cmd/engine-test/main.go <file.ai>")
		fmt.Println("       go run cmd/engine-test/main.go -e '<expr>'")
		os.Exit(1)
	}

	if os.Args[1] == "-e" && len(os.Args) >= 3 {
		result, err := runner.RunExpr(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
		return
	}

	err := runner.Run(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
