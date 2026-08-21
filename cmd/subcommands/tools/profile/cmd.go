package profile

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"aiki/engine"
	"aiki/engine/runner"
)

func Run(args []string) int {
	fs := flag.NewFlagSet("profile", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	countsOnly := fs.Bool("counts", false, "show semantic totals without source-site attribution")
	cpuPath := fs.String("cpu", "", "write a correlated Go CPU profile")
	allocsPath := fs.String("allocs", "", "write a Go allocation profile (pprof labels do not apply)")
	tracePath := fs.String("trace", "", "write a Go runtime trace for the measured evaluation")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki profile [-counts] [-cpu file] [-allocs file] [-trace file] file.ai")
		return 2
	}
	run, err := runner.RunProfileDetailed(fs.Arg(0), runner.ProfileOptions{
		Attributed:    !*countsOnly,
		CPUProfile:    *cpuPath,
		AllocsProfile: *allocsPath,
		TraceFile:     *tracePath,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "profile:", err)
		return 1
	}
	printMeasurement(run.Semantic, !*countsOnly)
	printSubstrate(run.Substrate)
	if *cpuPath != "" {
		fmt.Printf("  cpu profile  %s (labels: aiki_layer, aiki_function, aiki_file, aiki_line, aiki_primitive)\n", *cpuPath)
	}
	if *allocsPath != "" {
		fmt.Printf("  alloc profile %s (Go allocation sites; not Aiki-label correlated)\n", *allocsPath)
	}
	if *tracePath != "" {
		fmt.Printf("  runtime trace %s\n", *tracePath)
	}
	return 0
}

func printMeasurement(m engine.SemanticMeasurement, showSites bool) {
	c := m.Counts
	fmt.Println("Aiki semantic work")
	rows := []struct {
		name string
		n    int64
	}{
		{"arithmetic", c.Arithmetic}, {"comparison", c.Comparison},
		{"call", c.Call}, {"iteration", c.Iteration}, {"index", c.Index},
		{"send", c.Send}, {"receive", c.Receive},
		{"store_read", c.StoreRead}, {"store_write", c.StoreWrite},
	}
	for _, row := range rows {
		fmt.Printf("  %-12s %d\n", row.name, row.n)
	}
	printNumberRealization("Number arithmetic realization", m.Numbers, true)
	printNumberRealization("Number call-return realization", m.CallNumbers, false)
	if !showSites || len(m.Sites) == 0 {
		return
	}
	fmt.Println("\nAiki source attribution")
	for _, sc := range m.Sites {
		where := fmt.Sprintf("%s:%d:%d", sc.Site.File, sc.Site.Line, sc.Site.Col)
		context := ""
		if sc.Site.Function != "" {
			context = " in " + sc.Site.Function
		}
		if sc.Site.Detail != "" {
			context += " -> " + sc.Site.Detail
		}
		fmt.Printf("  %8d  %-12s %s%s\n", sc.Count, sc.Kind, where, context)
		if text := strings.TrimSpace(sc.Site.Source); text != "" {
			fmt.Printf("            %s\n", text)
		}
	}
}

func printNumberRealization(title string, n engine.NumberRealizationCounts, arithmetic bool) {
	total := n.ResultSmallInteger + n.ResultCompactRational + n.ResultBinaryCarrier + n.ResultBigRational
	if total == 0 {
		return
	}
	fmt.Printf("\n%s\n", title)
	fmt.Printf("  %-20s %d\n", "small_integer", n.ResultSmallInteger)
	fmt.Printf("  %-20s %d\n", "compact_rational", n.ResultCompactRational)
	fmt.Printf("  %-20s %d\n", "binary_carrier", n.ResultBinaryCarrier)
	fmt.Printf("  %-20s %d\n", "big_rational", n.ResultBigRational)
	if arithmetic {
		fmt.Printf("  %-20s %d\n", "binary_certified", n.BinaryCertified)
		fmt.Printf("  %-20s %d\n", "binary_fallback", n.BinaryFallback)
		fmt.Printf("  %-20s %d\n", "promoted_big", n.PromotedBigRational)
	}
}

func printSubstrate(s runner.SubstrateStats) {
	fmt.Println("\nGo substrate realization")
	fmt.Printf("  %-12s %s\n", "elapsed", s.Elapsed)
	fmt.Printf("  %-12s %d\n", "alloc_bytes", s.AllocBytes)
	fmt.Printf("  %-12s %d\n", "mallocs", s.Mallocs)
	fmt.Printf("  %-12s %d\n", "gc_cycles", s.GCs)
}
