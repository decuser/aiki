package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halBytesLength returns the length of a bytes value.
func halBytesLength(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("bytes_length: want 1 argument, got %d", len(args))
	}
	b, ok := args[0].(*value.Bytes)
	if !ok {
		return value.NewFault("bytes_length: expected bytes, got %s", args[0].Type())
	}
	return value.NewNumber(int64(len(b.Val)), 1)
}

// halBytesGet returns the byte at index i (0-255).
func halBytesGet(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bytes_get: want 2 arguments, got %d", len(args))
	}
	b, ok := args[0].(*value.Bytes)
	if !ok {
		return value.NewFault("bytes_get: expected bytes, got %s", args[0].Type())
	}
	idx, ok := args[1].(*value.Number)
	if !ok || !idx.Val.IsInt() {
		return value.NewFault("bytes_get: index must be integer")
	}
	i := int(idx.Val.Num().Int64())
	if i < 0 || i >= len(b.Val) {
		return value.NewFault("bytes_get: index %d out of bounds (length %d)", i, len(b.Val))
	}
	return value.NewNumber(int64(b.Val[i]), 1)
}

// halBytesSlice returns a slice of bytes from start to end.
func halBytesSlice(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("bytes_slice: want 3 arguments, got %d", len(args))
	}
	b, ok := args[0].(*value.Bytes)
	if !ok {
		return value.NewFault("bytes_slice: expected bytes, got %s", args[0].Type())
	}
	startNum, ok := args[1].(*value.Number)
	if !ok || !startNum.Val.IsInt() {
		return value.NewFault("bytes_slice: start must be integer")
	}
	endNum, ok := args[2].(*value.Number)
	if !ok || !endNum.Val.IsInt() {
		return value.NewFault("bytes_slice: end must be integer")
	}
	start := int(startNum.Val.Num().Int64())
	end := int(endNum.Val.Num().Int64())
	if start < 0 || end > len(b.Val) || start > end {
		return value.NewFault("bytes_slice: invalid range [%d:%d] for length %d", start, end, len(b.Val))
	}
	// Return new bytes (immutable)
	newBytes := make([]byte, end-start)
	copy(newBytes, b.Val[start:end])
	return &value.Bytes{Val: newBytes}
}

// halStrToBytes converts a string to UTF-8 bytes.
func halStrToBytes(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("str_to_bytes: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("str_to_bytes: expected string, got %s", args[0].Type())
	}
	return &value.Bytes{Val: []byte(s.Val)}
}

// halBytesToStr converts UTF-8 bytes to a string.
func halBytesToStr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("bytes_to_str: want 1 argument, got %d", len(args))
	}
	b, ok := args[0].(*value.Bytes)
	if !ok {
		return value.NewFault("bytes_to_str: expected bytes, got %s", args[0].Type())
	}
	return &value.String{Val: string(b.Val)}
}

// halBytesToStrPure converts a list of byte values to a string.
func halBytesToStrPure(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("bytes_to_str_pure: want 1 argument, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return value.NewFault("bytes_to_str_pure: expected list, got %s", args[0].Type())
	}
	bytes := make([]byte, len(list.Elements))
	for i, elem := range list.Elements {
		num, ok := elem.(*value.Number)
		if !ok || !num.Val.IsInt() {
			return value.NewFault("bytes_to_str_pure: element %d must be integer", i)
		}
		n := num.Val.Num().Int64()
		if n < 0 || n > 255 {
			return value.NewFault("bytes_to_str_pure: element %d out of range (0-255): %d", i, n)
		}
		bytes[i] = byte(n)
	}
	return &value.String{Val: string(bytes)}
}

// halBytesNew creates a bytes value from a list of integers (0-255).
func halBytesNew(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("bytes_new: want 1 argument, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return value.NewFault("bytes_new: expected list, got %s", args[0].Type())
	}
	result := make([]byte, len(list.Elements))
	for i, elem := range list.Elements {
		num, ok := elem.(*value.Number)
		if !ok || !num.Val.IsInt() {
			return value.NewFault("bytes_new: element %d must be integer", i)
		}
		n := num.Val.Num().Int64()
		if n < 0 || n > 255 {
			return value.NewFault("bytes_new: element %d out of range (0-255): %d", i, n)
		}
		result[i] = byte(n)
	}
	return &value.Bytes{Val: result}
}
