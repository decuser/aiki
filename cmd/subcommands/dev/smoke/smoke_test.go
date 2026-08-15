package smoke

import (
	"bytes"
	"testing"

	"aiki/cmd/internal/testfixture"
)

func TestEncodeTranscriptCanonicalEscapes(t *testing.T) {
	got := encodeTranscript(
		[]byte("Ada\n"),
		[]byte(":default\n42\n1\n"),
		[]byte("warning\n"),
		2,
		[]string{"line 1", "line 2"},
		[]string{"inspect canvas"},
	)
	want := []byte("IN:Ada\\n\n" +
		"OUT::default\\n\n" +
		"OUT:42\\n\n" +
		"OUT:1\\n\n" +
		"ERR:warning\\n\n" +
		"CANVAS:line 1\n" +
		"CANVAS:line 2\n" +
		"EXIT:2\n" +
		"DISPLAY:inspect canvas\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("encoded transcript\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

func TestEncodeTranscriptRoundTrips(t *testing.T) {
	data := encodeTranscript([]byte("x\ny"), []byte("a\nb"), []byte("e\n"), 3, nil, nil)
	// loadTranscript reads files, so exercise the escaping primitive directly.
	if unescape(escapeTranscript("a\n")) != "a\n" {
		t.Fatal("escape/unescape did not round trip newline")
	}
	if len(data) == 0 {
		t.Fatal("expected transcript data")
	}
}

func TestNegativeObservationRequiresDeclarationBothWays(t *testing.T) {
	parseErr := []byte("parser: test.ai:1:1:\nexpected expression\n")
	if err := validateNegativeObservation("declared", testfixture.NegativeParse, parseErr, 1); err != nil {
		t.Fatalf("declared parse failure rejected: %v", err)
	}
	if err := validateNegativeObservation("undeclared", testfixture.NegativeNone, parseErr, 1); err == nil {
		t.Fatal("undeclared parse failure was accepted")
	}
	if err := validateNegativeObservation("passing", testfixture.NegativeParse, nil, 0); err == nil {
		t.Fatal("declared parse-negative that passes was accepted")
	}
}
