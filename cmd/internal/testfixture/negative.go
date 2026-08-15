// Package testfixture recognizes declarations carried by intentionally
// negative Aiki test specimens. Declarations are textual because a parse-
// negative specimen, by definition, cannot be parsed as an Aiki program.
package testfixture

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type NegativeKind string

const (
	NegativeNone  NegativeKind = ""
	NegativeParse NegativeKind = "parse"
)

const markerPrefix = "# @negative "

// NegativeKindOf returns the declared negative-fixture kind for path.
// The marker must occur in the leading comment/blank-line header. Unknown
// kinds fail loudly so the declaration cannot silently become inert.
func NegativeKindOf(path string) (NegativeKind, error) {
	f, err := os.Open(path)
	if err != nil {
		return NegativeNone, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	lineNo := 0
	found := NegativeNone
	for s.Scan() {
		lineNo++
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		if !strings.HasPrefix(text, "#") {
			break
		}
		if !strings.HasPrefix(text, markerPrefix) {
			continue
		}
		if found != NegativeNone {
			return NegativeNone, fmt.Errorf("%s:%d: duplicate @negative declaration", path, lineNo)
		}
		if !strings.HasSuffix(path, "_smoke.ai") {
			return NegativeNone, fmt.Errorf("%s:%d: @negative declarations are only valid on *_smoke.ai fixtures", path, lineNo)
		}
		raw := strings.TrimSpace(strings.TrimPrefix(text, markerPrefix))
		switch NegativeKind(raw) {
		case NegativeParse:
			found = NegativeParse
		default:
			return NegativeNone, fmt.Errorf("%s:%d: unknown @negative kind %q", path, lineNo, raw)
		}
	}
	if err := s.Err(); err != nil {
		return NegativeNone, err
	}
	return found, nil
}

func IsParseNegative(path string) (bool, error) {
	kind, err := NegativeKindOf(path)
	return kind == NegativeParse, err
}
