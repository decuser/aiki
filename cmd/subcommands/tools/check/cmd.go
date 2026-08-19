package check

import (
	"flag"
	"fmt"
	"os"

	"aiki/engine/runtime/modules"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Run executes static source diagnostics that do not run the target program.
func Run(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffiUse := fs.Bool("ffi-use", false, "report direct and transitive FFI module imports")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if !*ffiUse || fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki check --ffi-use file.ai")
		return 2
	}

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "check:", err)
		return 1
	}
	home, _ := os.UserHomeDir()
	uses, err := modules.FFIUsage(g, fs.Arg(0), modules.DefaultModuleRoots(home))
	if err != nil {
		fmt.Fprintln(os.Stderr, "check:", err)
		return 1
	}
	if len(uses) == 0 {
		fmt.Println("FFI imports: none")
		return 0
	}
	fmt.Println("FFI imports:")
	for _, name := range uses {
		fmt.Printf("  %s\n", name)
	}
	return 0
}
