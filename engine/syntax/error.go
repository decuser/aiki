package syntax

import "aiki/engine"

// SourceError is a structured lexical or parse error. Error preserves the
// existing user-facing rendering while Pos/Kind/Message let services project
// diagnostics without scraping formatted text.
type SourceError struct {
	Kind     string
	Pos      engine.Position
	Message  string
	Rendered string
}

func (e *SourceError) Error() string {
	if e == nil {
		return ""
	}
	if e.Rendered != "" {
		return e.Rendered
	}
	return e.Message
}
