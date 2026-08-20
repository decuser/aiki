package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func fixedStoreNew(args []value.Value, kind, name string) value.Value {
	if len(args) != 1 {
		return value.NewFault("%s: want 1 argument, got %d", name, len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok || !n.Val.IsInt() || n.Val.Sign() < 0 || !n.Val.Num().IsInt64() {
		return value.NewFault("%s: size must be a non-negative integer", name)
	}
	sz := n.Val.Num().Int64()
	if sz > int64(^uint(0)>>1) {
		return value.NewFault("%s: size too large", name)
	}
	s := &value.FixedStore{Kind: kind}
	switch kind {
	case "byte":
		s.Bytes = make([]uint8, int(sz))
	case "word":
		s.Words = make([]uint16, int(sz))
	case "addr18":
		s.Addrs = make([]uint32, int(sz))
	case "counter":
		s.Counters = make([]uint64, int(sz))
	}
	return s
}
func halFixedStoreNewByte(args []value.Value, ctx *hal.EvalContext) value.Value {
	return fixedStoreNew(args, "byte", "store/ffi.new_byte")
}
func halFixedStoreNewWord(args []value.Value, ctx *hal.EvalContext) value.Value {
	return fixedStoreNew(args, "word", "store/ffi.new_word")
}
func halFixedStoreNewAddr18(args []value.Value, ctx *hal.EvalContext) value.Value {
	return fixedStoreNew(args, "addr18", "store/ffi.new_addr18")
}
func halFixedStoreNewCounter(args []value.Value, ctx *hal.EvalContext) value.Value {
	return fixedStoreNew(args, "counter", "store/ffi.new_counter")
}

func fixedStoreArg(args []value.Value, arity int, name string) (*value.FixedStore, int, value.Value) {
	if len(args) != arity {
		return nil, 0, value.NewFault("%s: want %d arguments, got %d", name, arity, len(args))
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok {
		return nil, 0, value.NewFault("%s: expected store/ffi", name)
	}
	n, ok := args[1].(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return nil, 0, value.NewFault("%s: index must be an integer", name)
	}
	i := n.Val.Num().Int64()
	if i < 0 || i >= int64(s.FixedLen()) {
		return nil, 0, value.NewFault("%s: index out of bounds", name)
	}
	return s, int(i), nil
}
func halFixedStoreGet(args []value.Value, ctx *hal.EvalContext) value.Value {
	s, i, f := fixedStoreArg(args, 2, "store/ffi.get")
	if f != nil {
		return f
	}
	v := s.FixedGet(i)
	switch s.Kind {
	case "byte":
		return newByte(v)
	case "word":
		return newWord(v)
	case "addr18":
		return newAddr18(v)
	default:
		return value.NewFault("store/ffi.get: unknown fixed store kind")
	}
}
func halFixedStoreSet(args []value.Value, ctx *hal.EvalContext) value.Value {
	s, i, f := fixedStoreArg(args, 3, "store/ffi.set")
	if f != nil {
		return f
	}
	o, ok := args[2].(*value.Opaque)
	if !ok || o.Provider != machineProvider || o.Kind != s.Kind {
		return value.NewFault("store/ffi.set: value does not match store kind")
	}
	s.FixedSet(i, o.Bits)
	return value.EMPTY
}
func counterStoreArg(args []value.Value, arity int, name string) (*value.FixedStore, int, value.Value) {
	s, i, f := fixedStoreArg(args, arity, name)
	if f != nil {
		return nil, 0, f
	}
	if s.Kind != "counter" {
		return nil, 0, value.NewFault("%s: expected counter store", name)
	}
	return s, i, nil
}
func counterNumber(v value.Value, name string) (uint64, value.Value) {
	n, ok := v.(*value.Number)
	if !ok || !n.Val.IsInt() || n.Val.Sign() < 0 || !n.Val.Num().IsUint64() {
		return 0, value.NewFault("%s: value must be a non-negative integer", name)
	}
	return n.Val.Num().Uint64(), nil
}
func halFixedStoreCounterGet(args []value.Value, ctx *hal.EvalContext) value.Value {
	s, i, f := counterStoreArg(args, 2, "store/ffi.counter_get")
	if f != nil {
		return f
	}
	v := s.CounterGet(i)
	if v > uint64(^uint64(0)>>1) {
		return value.NewFault("store/ffi.counter_get: value exceeds Aiki integer boundary")
	}
	return value.NewNumber(int64(v), 1)
}
func halFixedStoreCounterSet(args []value.Value, ctx *hal.EvalContext) value.Value {
	s, i, f := counterStoreArg(args, 3, "store/ffi.counter_set")
	if f != nil {
		return f
	}
	v, f := counterNumber(args[2], "store/ffi.counter_set")
	if f != nil {
		return f
	}
	s.CounterSet(i, v)
	return value.TRUE
}
func halFixedStoreCounterAdd(args []value.Value, ctx *hal.EvalContext) value.Value {
	s, i, f := counterStoreArg(args, 3, "store/ffi.counter_add")
	if f != nil {
		return f
	}
	delta, f := counterNumber(args[2], "store/ffi.counter_add")
	if f != nil {
		return f
	}
	before := s.CounterGet(i)
	if ^uint64(0)-before < delta {
		return value.NewFault("store/ffi.counter_add: counter overflow")
	}
	s.CounterAdd(i, delta)
	return value.TRUE
}

func halFixedStoreLength(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("store/ffi.length: want 1 argument")
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok {
		return value.NewFault("store/ffi.length: expected store/ffi")
	}
	return value.NewNumber(int64(s.FixedLen()), 1)
}
func halFixedStoreSnapshot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 || len(args) > 2 {
		return value.NewFault("store/ffi.snapshot: want 1 or 2 arguments")
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok {
		return value.NewFault("store/ffi.snapshot: expected store/ffi")
	}
	count := s.FixedLen()
	if len(args) == 2 {
		n, ok := args[1].(*value.Number)
		if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
			return value.NewFault("store/ffi.snapshot: count must be integer")
		}
		x := n.Val.Num().Int64()
		if x < 0 || x > int64(count) {
			return value.NewFault("store/ffi.snapshot: invalid count")
		}
		count = int(x)
	}
	out := make([]value.Value, count)
	if s.Kind == "counter" {
		for i := 0; i < count; i++ {
			v := s.CounterGet(i)
			if v > uint64(^uint64(0)>>1) {
				return value.NewFault("store/ffi.snapshot: counter exceeds Aiki integer boundary")
			}
			out[i] = value.NewNumber(int64(v), 1)
		}
		return &value.List{Elements: out}
	}
	raw := s.FixedSnapshot(count)
	for i, v := range raw {
		out[i] = value.NewNumber(int64(v), 1)
	}
	return &value.List{Elements: out}
}

// Direct physical-memory helpers keep byte/word values inside the fixed-width
// provider domain and avoid converting an 18-bit bus address to an Aiki index.
func halFixedStoreWordReadAddr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("store/ffi.word_read_addr: want 2 arguments")
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok || s.Kind != "word" {
		return value.NewFault("store/ffi.word_read_addr: expected word store")
	}
	a, f := machineValue(args[1], "addr18", "store/ffi.word_read_addr")
	if f != nil {
		return f
	}
	if a&1 != 0 {
		return value.NewFault("store/ffi.word_read_addr: odd address")
	}
	i := int(a >> 1)
	if i >= s.FixedLen() {
		return value.NewFault("store/ffi.word_read_addr: address out of bounds")
	}
	return newWord(s.FixedGet(i))
}
func halFixedStoreWordWriteAddr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("store/ffi.word_write_addr: want 3 arguments")
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok || s.Kind != "word" {
		return value.NewFault("store/ffi.word_write_addr: expected word store")
	}
	a, f := machineValue(args[1], "addr18", "store/ffi.word_write_addr")
	if f != nil {
		return f
	}
	if a&1 != 0 {
		return value.NewFault("store/ffi.word_write_addr: odd address")
	}
	w, f := machineValue(args[2], "word", "store/ffi.word_write_addr")
	if f != nil {
		return f
	}
	i := int(a >> 1)
	if i >= s.FixedLen() {
		return value.NewFault("store/ffi.word_write_addr: address out of bounds")
	}
	s.FixedSet(i, w)
	return value.TRUE
}
func halFixedStoreByteReadAddr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("store/ffi.byte_read_addr: want 2 arguments")
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok || s.Kind != "word" {
		return value.NewFault("store/ffi.byte_read_addr: expected word store")
	}
	a, f := machineValue(args[1], "addr18", "store/ffi.byte_read_addr")
	if f != nil {
		return f
	}
	i := int(a >> 1)
	if i >= s.FixedLen() {
		return value.NewFault("store/ffi.byte_read_addr: address out of bounds")
	}
	w := s.FixedGet(i)
	if a&1 != 0 {
		w >>= 8
	}
	return newByte(w)
}
func halFixedStoreByteWriteAddr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("store/ffi.byte_write_addr: want 3 arguments")
	}
	s, ok := args[0].(*value.FixedStore)
	if !ok || s.Kind != "word" {
		return value.NewFault("store/ffi.byte_write_addr: expected word store")
	}
	a, f := machineValue(args[1], "addr18", "store/ffi.byte_write_addr")
	if f != nil {
		return f
	}
	b, f := machineValue(args[2], "byte", "store/ffi.byte_write_addr")
	if f != nil {
		return f
	}
	i := int(a >> 1)
	if i >= s.FixedLen() {
		return value.NewFault("store/ffi.byte_write_addr: address out of bounds")
	}
	w := s.FixedGet(i)
	if a&1 == 0 {
		w = (w & 0xff00) | b
	} else {
		w = (w & 0x00ff) | (b << 8)
	}
	s.FixedSet(i, w)
	return value.TRUE
}
