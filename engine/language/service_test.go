package language

import (
	"reflect"
	"strings"
	"testing"

	languageobserve "aiki/engine/language/observe"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func testService(t *testing.T) *Service {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	return NewService(g)
}

func TestDiagnosticsLexErrorStructured(t *testing.T) {
	s := testService(t)
	d := s.Diagnostics(Document{ID: "mem", Path: "test.ai", Source: "let x = `\n"}, value.ScopeUser)
	if len(d) != 1 || d[0].Source != "lex" || d[0].Pos.Line != 1 || d[0].Pos.Col != 9 || !strings.Contains(d[0].Message, "unexpected character") {
		t.Fatalf("diagnostics = %#v", d)
	}
}

func TestDiagnosticsParseErrorStructured(t *testing.T) {
	s := testService(t)
	d := s.Diagnostics(Document{Path: "test.ai", Source: "let x =\n"}, value.ScopeUser)
	if len(d) != 1 || d[0].Source != "parse" || d[0].Pos.Line < 1 || d[0].Severity != "error" {
		t.Fatalf("diagnostics = %#v", d)
	}
}

func TestDiagnosticsStructural(t *testing.T) {
	s := testService(t)
	d := s.Diagnostics(Document{Path: "test.ai", Source: "let x = missing\n"}, value.ScopeUser)
	if len(d) != 1 || d[0].Source != "lint" || d[0].Severity != "error" || !strings.Contains(d[0].Message, "undefined") {
		t.Fatalf("diagnostics = %#v", d)
	}
}

type recordingObserver struct{ events []languageobserve.Event }

func (o *recordingObserver) ObserveLanguage(e languageobserve.Event) { o.events = append(o.events, e) }

type recordingProbe struct {
	counts map[languageobserve.Metric]int64
}

func (p *recordingProbe) HitLanguage(m languageobserve.Metric, n int64) { p.counts[m] += n }

func TestObservationAndProbeAreBehaviorNeutral(t *testing.T) {
	s := testService(t)
	doc := Document{ID: "doc-1", Path: "test.ai", Source: "let x = missing\n"}
	want := s.Diagnostics(doc, value.ScopeUser)
	o := &recordingObserver{}
	p := &recordingProbe{counts: make(map[languageobserve.Metric]int64)}
	s.SetObserver(o)
	s.SetProbe(p)
	got := s.Diagnostics(doc, value.ScopeUser)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("instrumented diagnostics changed: got %#v want %#v", got, want)
	}
	if len(o.events) < 2 || p.counts[languageobserve.MetricDiagnosticsRequest] != 1 || p.counts[languageobserve.MetricAnalysisRun] != 1 || p.counts[languageobserve.MetricDiagnostic] != 1 {
		t.Fatalf("events=%v counts=%v", o.events, p.counts)
	}
}
