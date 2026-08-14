package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"runtime/pprof"
	"runtime/trace"

	"aiki/cmd/subcommands/ux/repl"
	"aiki/engine/runner"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"

	aikidebug "aiki/cmd/subcommands/dev/debug"
	aikienginesmoke "aiki/cmd/subcommands/dev/enginesmoke"
	aikismoke "aiki/cmd/subcommands/dev/smoke"
	aikitest "aiki/cmd/subcommands/dev/test"
	aikifmt "aiki/cmd/subcommands/tools/fmt"
	aikilint "aiki/cmd/subcommands/tools/lint"
	aikiprofile "aiki/cmd/subcommands/tools/profile"
)

func main() {
	repl.AppVersion = Version

	// Check for subcommands before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Println(Version)
			return
		case "fmt":
			aikifmt.Run(os.Args[2:])
			return
		case "lint":
			aikilint.Run(os.Args[2:])
			return
		case "profile":
			os.Exit(aikiprofile.Run(os.Args[2:]))
		case "smoke":
			aikismoke.Run(os.Args[2:])
			return
		case "enginesmoke":
			aikienginesmoke.Run(os.Args[2:])
			return
		case "debug":
			aikidebug.Run(os.Args[2:])
			return
		case "test":
			os.Exit(aikitest.Run(os.Args[2:]))
		}
	}

	opts := parseOptions()
	stopProfiling := startProfiling(opts)
	if stopProfiling != nil {
		defer stopProfiling()
	}
	if opts.Canvas {
		runCanvasChild(opts)
		return
	}

	if opts.Expr != "" {
		runExpr(opts.Expr, opts)
		return
	}

	if flag.NArg() == 0 {
		startREPL(opts)
	} else {
		runFile(flag.Arg(0), opts)
	}
}

func startProfiling(opts Options) func() {
	var stops []func()

	if opts.CPUProfile != "" {
		f, err := os.Create(opts.CPUProfile)
		if err == nil {
			if err := pprof.StartCPUProfile(f); err == nil {
				stops = append(stops, func() {
					pprof.StopCPUProfile()
					_ = f.Close()
				})
			} else {
				_ = f.Close()
			}
		}
	}

	if opts.TraceFile != "" {
		f, err := os.Create(opts.TraceFile)
		if err == nil {
			if err := trace.Start(f); err == nil {
				stops = append(stops, func() {
					trace.Stop()
					_ = f.Close()
				})
			} else {
				_ = f.Close()
			}
		}
	}

	if len(stops) == 0 {
		return nil
	}
	return func() {
		for i := len(stops) - 1; i >= 0; i-- {
			stops[i]()
		}
	}
}

func startREPL(opts Options) {
	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "user"}
	}

	fmt.Printf("%s %s\n", appName, Version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for help.\n\n")

	sess, err := repl.NewSession(os.Stdout, opts.Debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	sess.Run()
}

func runFile(filename string, opts Options) {
	err := runner.Run(filename)
	substrate.CloseAllCanvases()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runExpr(expr string, opts Options) {
	sess, err := runner.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result := sess.Eval(expr)

	if _, ok := result.(*value.ExitSignal); ok {
		return
	}

	if f, ok := result.(*value.Fault); ok {
		fmt.Fprintln(os.Stderr, f.Inspect())
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}
