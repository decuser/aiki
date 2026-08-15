package tags

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/language"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func Run(args []string) int {
	fs := flag.NewFlagSet("tags", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("o", "tags", "output tags file ('-' for stdout)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki tags [-o file] <file.ai|dir|./...> [more paths]")
		return 2
	}
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tags:", err)
		return 1
	}
	svc := language.NewService(g)
	files, err := expand(fs.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "tags:", err)
		return 1
	}
	var lines []string
	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tags:", err)
			return 1
		}
		syms, err := svc.Symbols(language.Document{Path: path, Source: string(b)})
		if err != nil {
			fmt.Fprintln(os.Stderr, "tags:", err)
			return 1
		}
		for _, sym := range syms {
			if !sym.TopLevel {
				continue
			}
			kind := "v"
			if sym.Kind == "shape" {
				kind = "s"
			}
			lines = append(lines, fmt.Sprintf("%s\t%s\t%d;\"\t%s", sym.Name, path, sym.Pos.Line, kind))
		}
	}
	sort.Strings(lines)
	text := strings.Join(lines, "\n")
	if len(lines) > 0 {
		text += "\n"
	}
	if *output == "-" {
		fmt.Print(text)
		return 0
	}
	if err := os.WriteFile(*output, []byte(text), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "tags:", err)
		return 1
	}
	return 0
}
func expand(args []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, a := range args {
		if a == "./..." || a == "..." {
			a = "."
		}
		st, err := os.Stat(a)
		if err != nil {
			return nil, err
		}
		if !st.IsDir() {
			if strings.HasSuffix(a, ".ai") {
				add(filepath.Clean(a))
			}
			continue
		}
		err = filepath.WalkDir(a, func(p string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(p, ".ai") {
				add(filepath.Clean(p))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}
