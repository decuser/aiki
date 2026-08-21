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
	printCallRealization(m.Calls)
	printListRealization(m.Lists)
	printEnvRealization(m.Envs)
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

func printCallRealization(c engine.CallRealizationCounts) {
	total := c.UserEntry + c.Substrate + c.TailReuse + c.TailEnvReuse
	if total == 0 {
		return
	}
	fmt.Println("\nCall realization")
	fmt.Printf("  %-20s %d\n", "user_entry", c.UserEntry)
	fmt.Printf("  %-20s %d\n", "substrate", c.Substrate)
	fmt.Printf("  %-20s %d\n", "tail_reuse", c.TailReuse)
	fmt.Printf("  %-20s %d\n", "tail_env_reuse", c.TailEnvReuse)
	fmt.Printf("  %-20s %d\n", "arg_arity_0", c.ArgArity0)
	fmt.Printf("  %-20s %d\n", "arg_arity_1", c.ArgArity1)
	fmt.Printf("  %-20s %d\n", "arg_arity_2", c.ArgArity2)
	fmt.Printf("  %-20s %d\n", "arg_arity_3", c.ArgArity3)
	fmt.Printf("  %-20s %d\n", "arg_arity_4", c.ArgArity4)
	fmt.Printf("  %-20s %d\n", "arg_arity_5_plus", c.ArgArity5Plus)
	fmt.Printf("  %-20s %d\n", "args_evaluated", c.ArgsEvaluated)
	fmt.Printf("  %-20s %d\n", "arg_frame_new", c.ArgFrameNew)
	fmt.Printf("  %-20s %d\n", "arg_frame_reused", c.ArgFrameReused)
	fmt.Printf("  %-20s %d\n", "arg_frame_promoted", c.ArgFramePromoted)
	fmt.Printf("  %-20s %d\n", "arg_durable", c.ArgDurable)
	fmt.Printf("  %-20s %d\n", "arg_tail_transfer", c.ArgTailTransfer)
}

func printListRealization(c engine.ListRealizationCounts) {
	total := c.FrontierPromoted + c.FrontierExtended + c.FrontierForked
	if total == 0 {
		return
	}
	fmt.Println("\nList realization")
	fmt.Printf("  %-24s %d\n", "frontier_promoted", c.FrontierPromoted)
	fmt.Printf("  %-24s %d\n", "frontier_extended", c.FrontierExtended)
	fmt.Printf("  %-24s %d\n", "frontier_grown", c.FrontierGrown)
	fmt.Printf("  %-24s %d\n", "frontier_forked", c.FrontierForked)
	fmt.Printf("  %-24s %d\n", "elements_copied", c.ElementsCopied)
	fmt.Printf("  %-24s %d\n", "backing_slots_allocated", c.BackingSlotsAllocated)
}

func printEnvRealization(c engine.EnvRealizationCounts) {
	total := c.PhysicalCall + c.PhysicalEnclosed + c.PhysicalIsolated + c.LogicalCall
	if total == 0 {
		return
	}

	call0 := c.LogicalCall - c.CallReached1
	call1 := c.CallReached1 - c.CallReached2
	call2 := c.CallReached2 - c.CallReached3
	call34 := c.CallReached3 - c.CallReached5

	enclosed0 := c.PhysicalEnclosed - c.EnclosedReached1
	if enclosed0 < 0 {
		enclosed0 = 0
	}
	enclosed1 := c.EnclosedReached1 - c.EnclosedReached2
	enclosed2 := c.EnclosedReached2 - c.EnclosedReached3
	enclosed34 := c.EnclosedReached3 - c.EnclosedReached5

	fmt.Println("\nEnvironment realization")
	fmt.Printf("  %-28s %d\n", "physical_call", c.PhysicalCall)
	fmt.Printf("  %-28s %d\n", "logical_call", c.LogicalCall)
	fmt.Printf("  %-28s %d\n", "physical_enclosed", c.PhysicalEnclosed)
	fmt.Printf("  %-28s %d\n", "physical_isolated", c.PhysicalIsolated)

	fmt.Printf("  %-28s %d\n", "call_local_max_0", call0)
	fmt.Printf("  %-28s %d\n", "call_local_max_1", call1)
	fmt.Printf("  %-28s %d\n", "call_local_max_2", call2)
	fmt.Printf("  %-28s %d\n", "call_local_max_3_4", call34)
	fmt.Printf("  %-28s %d\n", "call_local_max_5_plus", c.CallReached5)

	fmt.Printf("  %-28s %d\n", "enclosed_local_max_0", enclosed0)
	fmt.Printf("  %-28s %d\n", "enclosed_local_max_1", enclosed1)
	fmt.Printf("  %-28s %d\n", "enclosed_local_max_2", enclosed2)
	fmt.Printf("  %-28s %d\n", "enclosed_local_max_3_4", enclosed34)
	fmt.Printf("  %-28s %d\n", "enclosed_local_max_5_plus", c.EnclosedReached5)

	fmt.Printf("  %-28s %d\n", "call_compact_allocations", c.CallCompactAllocations)
	fmt.Printf("  %-28s %d\n", "enclosed_compact_allocations", c.EnclosedCompactAllocations)
	fmt.Printf("  %-28s %d\n", "call_map_promotions", c.CallMapPromotions)
	fmt.Printf("  %-28s %d\n", "enclosed_map_promotions", c.EnclosedMapPromotions)
	fmt.Printf("  %-28s %d\n", "call_local_new", c.CallLocalNew)
	fmt.Printf("  %-28s %d\n", "call_local_update", c.CallLocalUpdate)
	fmt.Printf("  %-28s %d\n", "enclosed_local_new", c.EnclosedLocalNew)
	fmt.Printf("  %-28s %d\n", "enclosed_local_update", c.EnclosedLocalUpdate)
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
