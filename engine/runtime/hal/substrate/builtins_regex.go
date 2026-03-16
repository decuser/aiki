package substrate

import (
	"regexp"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halRegexMatch returns true if the pattern matches the string.
func halRegexMatch(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_regex_match: want 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_regex_match: expected string as first argument, got %s", args[0].Type())
	}
	p, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_regex_match: expected string pattern as second argument, got %s", args[1].Type())
	}

	re, err := regexp.Compile(p.Val)
	if err != nil {
		return value.NewShapedError("regex", "invalid pattern: %s", err)
	}

	if re.MatchString(s.Val) {
		return value.TRUE
	}
	return value.FALSE
}

// halRegexFind finds the first match of pattern in string.
// Returns [@match, start, end, text] or [@error, :regex, reason].
func halRegexFind(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_regex_find: want 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_regex_find: expected string as first argument, got %s", args[0].Type())
	}
	p, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_regex_find: expected string pattern as second argument, got %s", args[1].Type())
	}

	re, err := regexp.Compile(p.Val)
	if err != nil {
		return value.NewShapedError("regex", "invalid pattern: %s", err)
	}

	loc := re.FindStringIndex(s.Val)
	if loc == nil {
		return value.NewShapedError("regex", "no match")
	}

	return &value.List{
		Shape: "match",
		Elements: []value.Value{
			value.NewNumber(int64(loc[0]), 1),
			value.NewNumber(int64(loc[1]), 1),
			&value.String{Val: s.Val[loc[0]:loc[1]]},
		},
	}
}

// halRegexFindAll finds all matches of pattern in string.
// Returns a list of [@match, start, end, text].
func halRegexFindAll(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_regex_find_all: want 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_regex_find_all: expected string as first argument, got %s", args[0].Type())
	}
	p, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_regex_find_all: expected string pattern as second argument, got %s", args[1].Type())
	}

	re, err := regexp.Compile(p.Val)
	if err != nil {
		return value.NewShapedError("regex", "invalid pattern: %s", err)
	}

	locs := re.FindAllStringIndex(s.Val, -1)
	if locs == nil {
		return &value.List{Elements: []value.Value{}}
	}

	matches := make([]value.Value, len(locs))
	for i, loc := range locs {
		matches[i] = &value.List{
			Shape: "match",
			Elements: []value.Value{
				value.NewNumber(int64(loc[0]), 1),
				value.NewNumber(int64(loc[1]), 1),
				&value.String{Val: s.Val[loc[0]:loc[1]]},
			},
		}
	}
	return &value.List{Elements: matches}
}

// halRegexReplace replaces all matches of pattern with replacement.
func halRegexReplace(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("_regex_replace: want 3 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_regex_replace: expected string as first argument, got %s", args[0].Type())
	}
	p, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_regex_replace: expected string pattern as second argument, got %s", args[1].Type())
	}
	r, ok := args[2].(*value.String)
	if !ok {
		return value.NewFault("_regex_replace: expected string replacement as third argument, got %s", args[2].Type())
	}

	re, err := regexp.Compile(p.Val)
	if err != nil {
		return value.NewShapedError("regex", "invalid pattern: %s", err)
	}

	return &value.String{Val: re.ReplaceAllString(s.Val, r.Val)}
}

// halRegexReplaceFirst replaces only the first match of pattern with replacement.
func halRegexReplaceFirst(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("_regex_replace_first: want 3 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_regex_replace_first: expected string as first argument, got %s", args[0].Type())
	}
	p, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_regex_replace_first: expected string pattern as second argument, got %s", args[1].Type())
	}
	r, ok := args[2].(*value.String)
	if !ok {
		return value.NewFault("_regex_replace_first: expected string replacement as third argument, got %s", args[2].Type())
	}

	re, err := regexp.Compile(p.Val)
	if err != nil {
		return value.NewShapedError("regex", "invalid pattern: %s", err)
	}

	// Replace only the first match
	loc := re.FindStringIndex(s.Val)
	if loc == nil {
		return &value.String{Val: s.Val}
	}
	result := s.Val[:loc[0]] + r.Val + s.Val[loc[1]:]
	return &value.String{Val: result}
}

// halRegexSplit splits string by pattern.
func halRegexSplit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_regex_split: want 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_regex_split: expected string as first argument, got %s", args[0].Type())
	}
	p, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_regex_split: expected string pattern as second argument, got %s", args[1].Type())
	}

	re, err := regexp.Compile(p.Val)
	if err != nil {
		return value.NewShapedError("regex", "invalid pattern: %s", err)
	}

	parts := re.Split(s.Val, -1)
	elements := make([]value.Value, len(parts))
	for i, part := range parts {
		elements[i] = &value.String{Val: part}
	}
	return &value.List{Elements: elements}
}
