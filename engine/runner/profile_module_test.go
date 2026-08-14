package runner

import (
	"os"
	"path/filepath"
	"testing"

	"aiki/engine/semantics/value"
)

func withRepoRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(wd, "../.."))
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func TestProfileModuleCounts(t *testing.T) {
	withRepoRoot(t)
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	v := s.Eval(`
let profile = import("profile")
let work = () {
	1 + 2
}
profile.counts(work)
`)
	list, ok := v.(*value.List)
	if !ok {
		t.Fatalf("expected counts list, got %T: %v", v, v)
	}
	if len(list.Elements) != 9 {
		t.Fatalf("expected 9 count entries, got %d", len(list.Elements))
	}
	if list.Elements[0].Inspect() != "[:arithmetic, 1]" {
		t.Fatalf("arithmetic count: got %s", list.Elements[0].Inspect())
	}
	if list.Elements[2].Inspect() != "[:call, 1]" {
		t.Fatalf("call count: got %s", list.Elements[2].Inspect())
	}
}

func TestProfileModuleMeasureHasSites(t *testing.T) {
	withRepoRoot(t)
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	v := s.Eval(`
let profile = import("profile")
let work = () {
	1 + 2
}
profile.measure(work)
`)
	measured, ok := v.(*value.List)
	if !ok || len(measured.Elements) != 2 {
		t.Fatalf("expected [counts, sites], got %T: %v", v, v)
	}
	sites, ok := measured.Elements[1].(*value.List)
	if !ok || len(sites.Elements) == 0 {
		t.Fatalf("expected attributed sites, got %v", measured.Elements[1])
	}
}

func TestRunProfileAttributesSource(t *testing.T) {
	withRepoRoot(t)
	path := filepath.Join(".", ".profile_test_temp.ai")
	source := "let x = 1 + 2\nx < 4\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	m, err := RunProfile(path, true)
	if err != nil {
		t.Fatalf("RunProfile: %v", err)
	}
	if m.Counts.Arithmetic != 1 || m.Counts.Comparison != 1 {
		t.Fatalf("counts: arithmetic/comparison = %d/%d", m.Counts.Arithmetic, m.Counts.Comparison)
	}
	if len(m.Sites) < 2 {
		t.Fatalf("expected at least two attributed sites, got %d", len(m.Sites))
	}
	found := false
	for _, site := range m.Sites {
		if site.Kind == "arithmetic" && site.Site.Line == 1 && site.Site.Source == "let x = 1 + 2" {
			found = true
		}
	}
	if !found {
		t.Fatalf("did not find arithmetic source attribution: %#v", m.Sites)
	}
}

func TestRunProfileDetailedReportsSubstrateAndCPUArtifact(t *testing.T) {
	withRepoRoot(t)
	path := filepath.Join(".", ".profile_substrate_test.ai")
	cpu := filepath.Join(t.TempDir(), "cpu.pprof")
	source := "let x = 0\nwhile x < 2000 {\n\tx = x + 1\n}\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	run, err := RunProfileDetailed(path, ProfileOptions{Attributed: true, CPUProfile: cpu})
	if err != nil {
		t.Fatalf("RunProfileDetailed: %v", err)
	}
	if run.Semantic.Counts.Iteration != 2000 {
		t.Fatalf("iterations: expected 2000, got %d", run.Semantic.Counts.Iteration)
	}
	if run.Substrate.Elapsed <= 0 || run.Substrate.Mallocs == 0 {
		t.Fatalf("substrate stats not populated: %#v", run.Substrate)
	}
	info, err := os.Stat(cpu)
	if err != nil {
		t.Fatalf("cpu profile: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("cpu profile is empty")
	}
}

func TestRunProfileAttributesFunctionAndCallDetail(t *testing.T) {
	withRepoRoot(t)
	path := filepath.Join(".", ".profile_attribution_test.ai")
	source := "let add_one = (x) {\n\treturn x + 1\n}\nadd_one(41)\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	m, err := RunProfile(path, true)
	if err != nil {
		t.Fatalf("RunProfile: %v", err)
	}

	var foundCall, foundArithmetic bool
	for _, site := range m.Sites {
		if site.Kind == "call" && site.Site.Line == 4 && site.Site.Detail == "add_one" {
			foundCall = true
		}
		if site.Kind == "arithmetic" && site.Site.Line == 2 && site.Site.Function == "add_one" && site.Site.Source == "\treturn x + 1" {
			foundArithmetic = true
		}
	}
	if !foundCall {
		t.Fatalf("missing top-level call attribution to add_one: %#v", m.Sites)
	}
	if !foundArithmetic {
		t.Fatalf("missing function-body arithmetic attribution: %#v", m.Sites)
	}
}

func TestRunProfileAggregatesRepeatedSourceSite(t *testing.T) {
	withRepoRoot(t)
	path := filepath.Join(".", ".profile_aggregation_test.ai")
	source := "let i = 0\nwhile i < 3 {\n\ti = i + 1\n}\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	m, err := RunProfile(path, true)
	if err != nil {
		t.Fatalf("RunProfile: %v", err)
	}
	for _, site := range m.Sites {
		if site.Kind == "arithmetic" && site.Site.Line == 3 {
			if site.Count != 3 {
				t.Fatalf("arithmetic site count: got %d, want 3", site.Count)
			}
			return
		}
	}
	t.Fatalf("missing aggregated arithmetic site: %#v", m.Sites)
}
