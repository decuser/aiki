package substrate

import (
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func trimAikiWhitespace(s string, left, right bool) string {
	runes := []rune(s)
	start := 0
	end := len(runes)

	isSpace := func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}

	if left {
		for start < end && isSpace(runes[start]) {
			start++
		}
	}
	if right {
		for end > start && isSpace(runes[end-1]) {
			end--
		}
	}
	return string(runes[start:end])
}

func stringArg(args []value.Value, at int, name string) (*value.String, *value.Fault) {
	if at >= len(args) {
		return nil, value.NewFault("%s: missing argument", name)
	}
	s, ok := args[at].(*value.String)
	if !ok {
		return nil, value.NewFault("%s: expected string, got %s", name, args[at].Type())
	}
	return s, nil
}

func stringIntArg(args []value.Value, at int, name, what string) (int, *value.Fault) {
	if at >= len(args) {
		return 0, value.NewFault("%s: missing %s", name, what)
	}
	n, ok := args[at].(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return 0, value.NewFault("%s: %s must be an integer", name, what)
	}
	return int(n.Val.Num().Int64()), nil
}

func halStringSubstring(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("string_substring: want 3 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_substring")
	if fault != nil {
		return fault
	}
	start, fault := stringIntArg(args, 1, "string_substring", "start")
	if fault != nil {
		return fault
	}
	end, fault := stringIntArg(args, 2, "string_substring", "end")
	if fault != nil {
		return fault
	}
	if start < 0 {
		return value.NewFault("string_substring: start %d out of bounds", start)
	}
	if end <= start {
		return &value.String{Val: ""}
	}
	runes := []rune(s.Val)
	if start >= len(runes) {
		return &value.String{Val: ""}
	}
	if end > len(runes) {
		end = len(runes)
	}
	return &value.String{Val: string(runes[start:end])}
}

func halStringSplit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_split: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_split")
	if fault != nil {
		return fault
	}
	delim, fault := stringArg(args, 1, "string_split")
	if fault != nil {
		return fault
	}
	if delim.Val == "" {
		return value.NewFault("string_split: delimiter must not be empty")
	}
	parts := strings.Split(s.Val, delim.Val)
	elems := make([]value.Value, len(parts))
	for i, part := range parts {
		elems[i] = &value.String{Val: part}
	}
	return &value.List{Elements: elems}
}

func stringRuneIndex(s, sub string, last bool) int {
	runes := []rune(s)
	needle := []rune(sub)
	if len(needle) == 0 {
		if last {
			return len(runes)
		}
		return 0
	}
	if len(needle) > len(runes) {
		return -1
	}
	start, stop, step := 0, len(runes)-len(needle), 1
	if last {
		start, stop, step = len(runes)-len(needle), 0, -1
	}
	for i := start; ; i += step {
		match := true
		for j := range needle {
			if runes[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
		if i == stop {
			break
		}
	}
	return -1
}

func stringLookupResult(i int) value.Value {
	if i < 0 {
		return value.NewShapedError("lookup", "not found")
	}
	return value.NewNumber(int64(i), 1)
}

func halStringIndexOf(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_index_of: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_index_of")
	if fault != nil {
		return fault
	}
	sub, fault := stringArg(args, 1, "string_index_of")
	if fault != nil {
		return fault
	}
	return stringLookupResult(stringRuneIndex(s.Val, sub.Val, false))
}

func halStringLastIndexOf(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_last_index_of: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_last_index_of")
	if fault != nil {
		return fault
	}
	sub, fault := stringArg(args, 1, "string_last_index_of")
	if fault != nil {
		return fault
	}
	return stringLookupResult(stringRuneIndex(s.Val, sub.Val, true))
}

func halStringContains(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_contains: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_contains")
	if fault != nil {
		return fault
	}
	sub, fault := stringArg(args, 1, "string_contains")
	if fault != nil {
		return fault
	}
	if strings.Contains(s.Val, sub.Val) {
		return value.TRUE
	}
	return value.FALSE
}

func halStringStartsWith(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_starts_with: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_starts_with")
	if fault != nil {
		return fault
	}
	prefix, fault := stringArg(args, 1, "string_starts_with")
	if fault != nil {
		return fault
	}
	if strings.HasPrefix(s.Val, prefix.Val) {
		return value.TRUE
	}
	return value.FALSE
}

func halStringEndsWith(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_ends_with: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_ends_with")
	if fault != nil {
		return fault
	}
	suffix, fault := stringArg(args, 1, "string_ends_with")
	if fault != nil {
		return fault
	}
	if strings.HasSuffix(s.Val, suffix.Val) {
		return value.TRUE
	}
	return value.FALSE
}

func halStringJoin(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_join: want 2 arguments, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return value.NewFault("string_join: expected list, got %s", args[0].Type())
	}
	delim, fault := stringArg(args, 1, "string_join")
	if fault != nil {
		return fault
	}
	parts := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		s, ok := elem.(*value.String)
		if !ok {
			return value.NewFault("string_join: element %d must be string", i)
		}
		parts[i] = s.Val
	}
	return &value.String{Val: strings.Join(parts, delim.Val)}
}

func halStringReplace(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("string_replace: want 3 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_replace")
	if fault != nil {
		return fault
	}
	old, fault := stringArg(args, 1, "string_replace")
	if fault != nil {
		return fault
	}
	newValue, fault := stringArg(args, 2, "string_replace")
	if fault != nil {
		return fault
	}
	if old.Val == "" {
		return value.NewFault("string_replace: old string must not be empty")
	}
	return &value.String{Val: strings.ReplaceAll(s.Val, old.Val, newValue.Val)}
}

func halStringReplaceFirst(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("string_replace_first: want 3 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_replace_first")
	if fault != nil {
		return fault
	}
	old, fault := stringArg(args, 1, "string_replace_first")
	if fault != nil {
		return fault
	}
	newValue, fault := stringArg(args, 2, "string_replace_first")
	if fault != nil {
		return fault
	}
	if old.Val == "" {
		return value.NewFault("string_replace_first: old string must not be empty")
	}
	return &value.String{Val: strings.Replace(s.Val, old.Val, newValue.Val, 1)}
}

func halStringRepeat(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_repeat: want 2 arguments, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_repeat")
	if fault != nil {
		return fault
	}
	n, fault := stringIntArg(args, 1, "string_repeat", "count")
	if fault != nil {
		return fault
	}
	if n <= 0 {
		return &value.String{Val: ""}
	}
	return &value.String{Val: strings.Repeat(s.Val, n)}
}

func halStringReverse(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("string_reverse: want 1 argument, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_reverse")
	if fault != nil {
		return fault
	}
	runes := []rune(s.Val)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return &value.String{Val: string(runes)}
}

func aikiWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func trimRunes(s string, left, right bool) string {
	runes := []rune(s)
	start, end := 0, len(runes)
	if left {
		for start < end && aikiWhitespace(runes[start]) {
			start++
		}
	}
	if right {
		for end > start && aikiWhitespace(runes[end-1]) {
			end--
		}
	}
	return string(runes[start:end])
}

func halStringTrim(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("string_trim: want 1 argument, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_trim")
	if fault != nil {
		return fault
	}
	return &value.String{Val: trimRunes(s.Val, true, true)}
}

func halStringTrimStart(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("string_trim_start: want 1 argument, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_trim_start")
	if fault != nil {
		return fault
	}
	return &value.String{Val: trimRunes(s.Val, true, false)}
}

func halStringTrimEnd(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("string_trim_end: want 1 argument, got %d", len(args))
	}
	s, fault := stringArg(args, 0, "string_trim_end")
	if fault != nil {
		return fault
	}
	return &value.String{Val: trimRunes(s.Val, false, true)}
}

func halStringCompare(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("string_compare: want 2 arguments, got %d", len(args))
	}
	a, fault := stringArg(args, 0, "string_compare")
	if fault != nil {
		return fault
	}
	b, fault := stringArg(args, 1, "string_compare")
	if fault != nil {
		return fault
	}
	ar, br := []rune(a.Val), []rune(b.Val)
	limit := len(ar)
	if len(br) < limit {
		limit = len(br)
	}
	for i := 0; i < limit; i++ {
		if ar[i] < br[i] {
			return value.NewNumber(-1, 1)
		}
		if ar[i] > br[i] {
			return value.NewNumber(1, 1)
		}
	}
	if len(ar) < len(br) {
		return value.NewNumber(-1, 1)
	}
	if len(ar) > len(br) {
		return value.NewNumber(1, 1)
	}
	return value.NewNumber(0, 1)
}
