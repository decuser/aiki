package runner

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/libpath"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// ProfileOptions controls one correlated semantic/substrate measurement.
type ProfileOptions struct {
	Attributed    bool
	CPUProfile    string
	AllocsProfile string
	TraceFile     string
	ProfileLabels bool
}

// SubstrateStats describes host work observed during the same evaluation
// interval as the semantic measurement. These are realization measurements,
// not Aiki semantic units.
type SubstrateStats struct {
	Elapsed    time.Duration
	AllocBytes uint64
	Mallocs    uint64
	GCs        uint32
}

// ProfileRun contains the two deliberately distinct views of one execution.
type ProfileRun struct {
	Semantic  engine.SemanticMeasurement
	Substrate SubstrateStats
}

// RunProfileDetailed evaluates filename once. Prelude loading, module registry
// setup, lexing, and parsing occur before the measured interval; the interval
// begins immediately before user AST evaluation.
func RunProfileDetailed(filename string, opts ProfileOptions) (ProfileRun, error) {
	var out ProfileRun
	source, err := os.ReadFile(filename)
	if err != nil {
		return out, fmt.Errorf("reading file: %w", err)
	}
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return out, fmt.Errorf("loading grammar: %w", err)
	}
	rt := substrate.NewGoRuntime()
	defer rt.CloseAllResources()
	if err := initModuleRegistry(g, rt); err != nil {
		return out, fmt.Errorf("initializing registry: %w", err)
	}

	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return out, fmt.Errorf("loading prelude: %w", err)
	}
	userScope := value.ScopeUser
	if libpath.IsBlessedLibPath(filename) {
		userScope = value.ScopePrelude
	}
	userEnv := value.NewEnclosedEnvWithScope(preludeEnv, userScope)
	userEnv.SetAuthority(rt.AuthorityForSource(filename))
	userEnv.SetFile(filename)
	userEnv.SetSource(string(source))

	lexer := syntax.NewLexer(g, filename, string(source), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return out, fmt.Errorf("lexer: %w", err)
	}
	parser := syntax.NewParser(g, tokens, string(source), nil)
	ast, err := parser.Parse()
	if err != nil {
		return out, fmt.Errorf("parser: %w", err)
	}

	var counters *evaluator.Counters
	if opts.Attributed {
		counters = evaluator.NewAttributedCounters()
	} else {
		counters = evaluator.NewCounters()
	}
	userEnv.SetSemanticProbe(counters)

	ev := evaluator.New(rt, nil)
	ev.SetGrammar(g)

	labels := opts.ProfileLabels || opts.CPUProfile != ""
	rt.SetProfileLabels(labels)

	stopCPU, err := startCPUProfile(opts.CPUProfile)
	if err != nil {
		return out, err
	}
	defer stopCPU()
	stopTrace, err := startRuntimeTrace(opts.TraceFile)
	if err != nil {
		return out, err
	}
	defer stopTrace()

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	var result value.Value
	rt.WithProfileLabels(engine.ProfileLabels{
		Layer:    "semantic",
		Function: "<main>",
		File:     filename,
	}, engine.ProfileLabels{}, func() {
		result = ev.Eval(ast, userEnv)
	})
	out.Substrate.Elapsed = time.Since(started)
	runtime.ReadMemStats(&after)

	out.Semantic = counters.Measurement()
	if err := writeRuntimeProfile("allocs", opts.AllocsProfile); err != nil {
		return out, err
	}
	out.Substrate.AllocBytes = after.TotalAlloc - before.TotalAlloc
	out.Substrate.Mallocs = after.Mallocs - before.Mallocs
	out.Substrate.GCs = after.NumGC - before.NumGC

	if fault, ok := result.(*value.Fault); ok {
		return out, fmt.Errorf("%s", fault.Inspect())
	}
	return out, nil
}

func startCPUProfile(path string) (func(), error) {
	if path == "" {
		return func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating cpu profile: %w", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("starting cpu profile: %w", err)
	}
	return func() {
		pprof.StopCPUProfile()
		_ = f.Close()
	}, nil
}

func startRuntimeTrace(path string) (func(), error) {
	if path == "" {
		return func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating runtime trace: %w", err)
	}
	if err := trace.Start(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("starting runtime trace: %w", err)
	}
	return func() {
		trace.Stop()
		_ = f.Close()
	}, nil
}

func writeRuntimeProfile(name, path string) error {
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s profile: %w", name, err)
	}
	defer f.Close()
	p := pprof.Lookup(name)
	if p == nil {
		return fmt.Errorf("runtime profile %q not available", name)
	}
	if err := p.WriteTo(f, 0); err != nil {
		return fmt.Errorf("writing %s profile: %w", name, err)
	}
	return nil
}
