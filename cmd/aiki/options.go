package main

import (
	"flag"
	"fmt"
	"os"

	"aiki/internal/version"
)

const appName = "aiki"

type Options struct {
	Version bool
	Debug   bool
	Expr    string
}

func parseOptions() Options {
	showVersion := flag.Bool("v", false, "show version and exit")
	debug := flag.Bool("d", false, "debug output")
	expr := flag.String("e", "", "evaluate expression and exit")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: aiki [options] [file.ai]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s\n", appName, version.Version)
		os.Exit(0)
	}

	return Options{
		Version: *showVersion,
		Debug:   *debug,
		Expr:    *expr,
	}
}
