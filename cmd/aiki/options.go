package main

import (
	"flag"
	"fmt"
	"os"
)

const appName = "aiki"

type Options struct {
	Version bool
	Debug   bool
	Expr    string
	Canvas  bool
	CanvasW int
	CanvasH int
}

func parseOptions() Options {
	showVersion := flag.Bool("v", false, "show version and exit")
	debug := flag.Bool("d", false, "debug output")
	expr := flag.String("e", "", "evaluate expression and exit")
	canvas := flag.Bool("canvas", false, "run as canvas child process")
	canvasW := flag.Int("canvasw", 0, "canvas child width")
	canvasH := flag.Int("canvash", 0, "canvas child height")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: aiki [options] [file.ai]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s\n", appName, Version)
		os.Exit(0)
	}

	return Options{
		Version: *showVersion,
		Debug:   *debug,
		Expr:    *expr,
		Canvas:  *canvas,
		CanvasW: *canvasW,
		CanvasH: *canvasH,
	}
}
