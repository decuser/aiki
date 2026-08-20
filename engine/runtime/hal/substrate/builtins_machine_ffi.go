package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"math/big"
)

const machineProvider = "machine/ffi"

func exactNonNegative(args []value.Value, name string, max uint64) (uint64, value.Value) {
	if len(args) != 1 {
		return 0, value.NewFault("%s: want 1 argument, got %d", name, len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok || !n.Val.IsInt() || n.Val.Sign() < 0 {
		return 0, value.NewFault("%s: expected non-negative integral number", name)
	}
	if !n.Val.Num().IsUint64() {
		return 0, value.NewFault("%s: value out of range", name)
	}
	v := n.Val.Num().Uint64()
	if v > max {
		return 0, value.NewFault("%s: value out of range", name)
	}
	return v, nil
}

func machineValue(v value.Value, kind, name string) (uint32, value.Value) {
	o, ok := v.(*value.Opaque)
	if !ok || o.Provider != machineProvider || o.Kind != kind {
		return 0, value.NewFault("%s: expected opaque machine value", name)
	}
	return o.Bits, nil
}

var machineBytes [1 << 8]value.Opaque
var machineWords [1 << 16]value.Opaque
var machineAddrs [1 << 18]value.Opaque

func init() {
	for i := range machineBytes {
		machineBytes[i] = value.Opaque{Provider: machineProvider, Kind: "byte", Bits: uint32(i)}
	}
	for i := range machineWords {
		machineWords[i] = value.Opaque{Provider: machineProvider, Kind: "word", Bits: uint32(i)}
	}
	for i := range machineAddrs {
		machineAddrs[i] = value.Opaque{Provider: machineProvider, Kind: "addr18", Bits: uint32(i)}
	}
}

// Fixed-width machine values are finite domains. Interning them avoids turning
// every register read, address calculation, bit operation, and memory fetch into
// a heap allocation while preserving the opaque Aiki boundary exactly.
func newWord(v uint32) value.Value   { return &machineWords[v&0xffff] }
func newByte(v uint32) value.Value   { return &machineBytes[v&0xff] }
func newAddr18(v uint32) value.Value { return &machineAddrs[v&0x3ffff] }

func halMachineWord(args []value.Value, ctx *hal.EvalContext) value.Value {
	v, f := exactNonNegative(args, "machine.word", 0xffff)
	if f != nil {
		return f
	}
	return newWord(uint32(v))
}
func halMachineByte(args []value.Value, ctx *hal.EvalContext) value.Value {
	v, f := exactNonNegative(args, "machine.byte", 0xff)
	if f != nil {
		return f
	}
	return newByte(uint32(v))
}
func halMachineAddr18(args []value.Value, ctx *hal.EvalContext) value.Value {
	v, f := exactNonNegative(args, "machine.addr18", 0x3ffff)
	if f != nil {
		return f
	}
	return newAddr18(uint32(v))
}
func halMachineToNumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.number: want 1 argument")
	}
	o, ok := args[0].(*value.Opaque)
	if !ok || o.Provider != machineProvider {
		return value.NewFault("machine.number: expected opaque machine value")
	}
	return &value.Number{Val: new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(o.Bits)))}
}
func halMachineSame(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.same: want 2 arguments")
	}
	a, aok := args[0].(*value.Opaque)
	b, bok := args[1].(*value.Opaque)
	if !aok || !bok || a.Provider != machineProvider || b.Provider != machineProvider {
		return value.NewFault("machine.same: expected opaque machine values")
	}
	if a.Kind == b.Kind && a.Bits == b.Bits {
		return value.TRUE
	}
	return value.FALSE
}

func wordBinary(args []value.Value, name string, op func(uint32, uint32) uint32) value.Value {
	if len(args) != 2 {
		return value.NewFault("%s: want 2 arguments", name)
	}
	a, f := machineValue(args[0], "word", name)
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "word", name)
	if f != nil {
		return f
	}
	return newWord(op(a, b))
}
func halMachineWordAdd(args []value.Value, ctx *hal.EvalContext) value.Value {
	return wordBinary(args, "machine.word_add", func(a, b uint32) uint32 { return a + b })
}
func halMachineWordSub(args []value.Value, ctx *hal.EvalContext) value.Value {
	return wordBinary(args, "machine.word_sub", func(a, b uint32) uint32 { return a - b })
}
func halMachineWordAnd(args []value.Value, ctx *hal.EvalContext) value.Value {
	return wordBinary(args, "machine.word_and", func(a, b uint32) uint32 { return a & b })
}
func halMachineWordOr(args []value.Value, ctx *hal.EvalContext) value.Value {
	return wordBinary(args, "machine.word_or", func(a, b uint32) uint32 { return a | b })
}
func halMachineWordXor(args []value.Value, ctx *hal.EvalContext) value.Value {
	return wordBinary(args, "machine.word_xor", func(a, b uint32) uint32 { return a ^ b })
}
func halMachineWordNot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.word_not: want 1 argument")
	}
	a, f := machineValue(args[0], "word", "machine.word_not")
	if f != nil {
		return f
	}
	return newWord(^a)
}
func halMachineWordShl(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_shl: want 2 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_shl")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:], "machine.word_shl", 63)
	if ff != nil {
		return ff
	}
	return newWord(a << uint(n))
}
func halMachineWordShr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_shr: want 2 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_shr")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:], "machine.word_shr", 63)
	if ff != nil {
		return ff
	}
	return newWord(a >> uint(n))
}
func halMachineWordExtract(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("machine.word_extract: want 3 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_extract")
	if f != nil {
		return f
	}
	start, ff := exactNonNegative(args[1:2], "machine.word_extract", 15)
	if ff != nil {
		return ff
	}
	width, ff := exactNonNegative(args[2:3], "machine.word_extract", 16)
	if ff != nil {
		return ff
	}
	if width == 0 || start+width > 16 {
		return value.NewFault("machine.word_extract: invalid field")
	}
	return newWord((a >> uint(start)) & ((1 << uint(width)) - 1))
}
func halMachineWordLT(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_lt: want 2 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_lt")
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "word", "machine.word_lt")
	if f != nil {
		return f
	}
	if a < b {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordGT(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_gt: want 2 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_gt")
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "word", "machine.word_gt")
	if f != nil {
		return f
	}
	if a > b {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordZero(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.word_zero: want 1 argument")
	}
	a, f := machineValue(args[0], "word", "machine.word_zero")
	if f != nil {
		return f
	}
	if a == 0 {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordSign(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.word_sign: want 1 argument")
	}
	a, f := machineValue(args[0], "word", "machine.word_sign")
	if f != nil {
		return f
	}
	if a&0x8000 != 0 {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordLowByte(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.low_byte: want 1 argument")
	}
	a, f := machineValue(args[0], "word", "machine.low_byte")
	if f != nil {
		return f
	}
	return newByte(a)
}
func halMachineWordHighByte(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.high_byte: want 1 argument")
	}
	a, f := machineValue(args[0], "word", "machine.high_byte")
	if f != nil {
		return f
	}
	return newByte(a >> 8)
}
func halMachineWordWithLowByte(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.with_low_byte: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.with_low_byte")
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "byte", "machine.with_low_byte")
	if f != nil {
		return f
	}
	return newWord((w & 0xff00) | b)
}
func halMachineWordWithHighByte(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.with_high_byte: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.with_high_byte")
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "byte", "machine.with_high_byte")
	if f != nil {
		return f
	}
	return newWord((w & 0x00ff) | (b << 8))
}
func halMachineByteToWord(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.byte_to_word: want 1 argument")
	}
	b, f := machineValue(args[0], "byte", "machine.byte_to_word")
	if f != nil {
		return f
	}
	return newWord(b)
}
func halMachineByteSignWord(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.byte_sign_word: want 1 argument")
	}
	b, f := machineValue(args[0], "byte", "machine.byte_sign_word")
	if f != nil {
		return f
	}
	if b&0x80 != 0 {
		return newWord(b | 0xff00)
	}
	return newWord(b)
}
func halMachineByteZero(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.byte_zero: want 1 argument")
	}
	b, f := machineValue(args[0], "byte", "machine.byte_zero")
	if f != nil {
		return f
	}
	if b == 0 {
		return value.TRUE
	}
	return value.FALSE
}

func halMachineAddrAdd(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.addr_add: want 2 arguments")
	}
	a, f := machineValue(args[0], "addr18", "machine.addr_add")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:], "machine.addr_add", 0x3ffff)
	if ff != nil {
		return ff
	}
	return newAddr18(a + uint32(n))
}
func halMachineAddrEven(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.addr_even: want 1 argument")
	}
	a, f := machineValue(args[0], "addr18", "machine.addr_even")
	if f != nil {
		return f
	}
	if a&1 == 0 {
		return value.TRUE
	}
	return value.FALSE
}

func halMachineWordField(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("machine.word_field: want 3 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_field")
	if f != nil {
		return f
	}
	start, ff := exactNonNegative(args[1:2], "machine.word_field", 15)
	if ff != nil {
		return ff
	}
	width, ff := exactNonNegative(args[2:3], "machine.word_field", 16)
	if ff != nil {
		return ff
	}
	if width == 0 || start+width > 16 {
		return value.NewFault("machine.word_field: invalid field")
	}
	v := (w >> uint(start)) & ((1 << uint(width)) - 1)
	return value.NewNumber(int64(v), 1)
}
func halMachineWordMask(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_mask: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_mask")
	if f != nil {
		return f
	}
	m, ff := exactNonNegative(args[1:2], "machine.word_mask", 0xffff)
	if ff != nil {
		return ff
	}
	return value.NewNumber(int64(w&uint32(m)), 1)
}
func halMachineWordEqNumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_eq_number: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_eq_number")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:2], "machine.word_eq_number", 0xffff)
	if ff != nil {
		return ff
	}
	if w == uint32(n) {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordLTNumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_lt_number: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_lt_number")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:2], "machine.word_lt_number", 0xffff)
	if ff != nil {
		return ff
	}
	if w < uint32(n) {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordGENumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_ge_number: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_ge_number")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:2], "machine.word_ge_number", 0xffff)
	if ff != nil {
		return ff
	}
	if w >= uint32(n) {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordAddSmall(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_add_small: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_add_small")
	if f != nil {
		return f
	}
	n, ok := args[1].(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return value.NewFault("machine.word_add_small: amount must be integer")
	}
	return newWord(uint32(int64(w) + n.Val.Num().Int64()))
}
func halMachineWordSubSmall(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_sub_small: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_sub_small")
	if f != nil {
		return f
	}
	n, ok := args[1].(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return value.NewFault("machine.word_sub_small: amount must be integer")
	}
	return newWord(uint32(int64(w) - n.Val.Num().Int64()))
}
func halMachineWordAddCarry(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_add_carry: want 2 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_add_carry")
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "word", "machine.word_add_carry")
	if f != nil {
		return f
	}
	if uint64(a)+uint64(b) > 0xffff {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordSubBorrow(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_sub_borrow: want 2 arguments")
	}
	a, f := machineValue(args[0], "word", "machine.word_sub_borrow")
	if f != nil {
		return f
	}
	b, f := machineValue(args[1], "word", "machine.word_sub_borrow")
	if f != nil {
		return f
	}
	if a < b {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordBit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_bit: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_bit")
	if f != nil {
		return f
	}
	p, ff := exactNonNegative(args[1:2], "machine.word_bit", 15)
	if ff != nil {
		return ff
	}
	if w&(1<<uint(p)) != 0 {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineWordSetBit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("machine.word_set_bit: want 3 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_set_bit")
	if f != nil {
		return f
	}
	p, ff := exactNonNegative(args[1:2], "machine.word_set_bit", 15)
	if ff != nil {
		return ff
	}
	b, ok := args[2].(*value.Boolean)
	if !ok {
		return value.NewFault("machine.word_set_bit: value must be boolean")
	}
	mask := uint32(1 << uint(p))
	if b.Val {
		w |= mask
	} else {
		w &= ^mask
	}
	return newWord(w)
}
func halMachineZero(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.zero: want 1 argument")
	}
	o, ok := args[0].(*value.Opaque)
	if !ok || o.Provider != machineProvider {
		return value.NewFault("machine.zero: expected opaque machine value")
	}
	if o.Bits == 0 {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineNegative(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.negative: want 1 argument")
	}
	o, ok := args[0].(*value.Opaque)
	if !ok || o.Provider != machineProvider {
		return value.NewFault("machine.negative: expected opaque machine value")
	}
	var mask uint32
	switch o.Kind {
	case "byte":
		mask = 0x80
	case "word":
		mask = 0x8000
	default:
		return value.NewFault("machine.negative: expected byte or word")
	}
	if o.Bits&mask != 0 {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineAddrFromWord(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.addr_from_word: want 1 argument")
	}
	w, f := machineValue(args[0], "word", "machine.addr_from_word")
	if f != nil {
		return f
	}
	return newAddr18(w)
}
func halMachineAddrAddOffset(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halMachineAddrAdd(args, ctx)
}
func halMachineAddrLTNumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.addr_lt_number: want 2 arguments")
	}
	a, f := machineValue(args[0], "addr18", "machine.addr_lt_number")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:2], "machine.addr_lt_number", 0x3ffff)
	if ff != nil {
		return ff
	}
	if a < uint32(n) {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineAddrGENumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.addr_ge_number: want 2 arguments")
	}
	a, f := machineValue(args[0], "addr18", "machine.addr_ge_number")
	if f != nil {
		return f
	}
	n, ff := exactNonNegative(args[1:2], "machine.addr_ge_number", 0x3ffff)
	if ff != nil {
		return ff
	}
	if a >= uint32(n) {
		return value.TRUE
	}
	return value.FALSE
}

func machineScalar(v value.Value, name string) (*value.Opaque, value.Value) {
	o, ok := v.(*value.Opaque)
	if !ok || o.Provider != machineProvider || (o.Kind != "word" && o.Kind != "byte") {
		return nil, value.NewFault("%s: expected opaque machine byte or word", name)
	}
	return o, nil
}
func machineMask(kind string) uint32 {
	if kind == "byte" {
		return 0xff
	}
	return 0xffff
}
func newSame(kind string, v uint32) value.Value {
	if kind == "byte" {
		return newByte(v)
	}
	return newWord(v)
}
func sameKindBinary(args []value.Value, name string, op func(uint32, uint32) uint32) value.Value {
	if len(args) != 2 {
		return value.NewFault("%s: want 2 arguments", name)
	}
	a, f := machineScalar(args[0], name)
	if f != nil {
		return f
	}
	b, f := machineScalar(args[1], name)
	if f != nil {
		return f
	}
	if a.Kind != b.Kind {
		return value.NewFault("%s: machine widths differ", name)
	}
	return newSame(a.Kind, op(a.Bits, b.Bits))
}
func halMachineAdd(args []value.Value, ctx *hal.EvalContext) value.Value {
	return sameKindBinary(args, "machine.add", func(a, b uint32) uint32 { return a + b })
}
func halMachineSub(args []value.Value, ctx *hal.EvalContext) value.Value {
	return sameKindBinary(args, "machine.sub", func(a, b uint32) uint32 { return a - b })
}
func halMachineAnd(args []value.Value, ctx *hal.EvalContext) value.Value {
	return sameKindBinary(args, "machine.and", func(a, b uint32) uint32 { return a & b })
}
func halMachineOr(args []value.Value, ctx *hal.EvalContext) value.Value {
	return sameKindBinary(args, "machine.or", func(a, b uint32) uint32 { return a | b })
}
func halMachineXor(args []value.Value, ctx *hal.EvalContext) value.Value {
	return sameKindBinary(args, "machine.xor", func(a, b uint32) uint32 { return a ^ b })
}
func halMachineNot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.not: want 1 argument")
	}
	a, f := machineScalar(args[0], "machine.not")
	if f != nil {
		return f
	}
	return newSame(a.Kind, ^a.Bits)
}
func halMachineAddSmall(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.add_small: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.add_small")
	if f != nil {
		return f
	}
	n, ok := args[1].(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return value.NewFault("machine.add_small: amount must be integer")
	}
	return newSame(a.Kind, uint32(int64(a.Bits)+n.Val.Num().Int64()))
}
func halMachineSubSmall(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.sub_small: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.sub_small")
	if f != nil {
		return f
	}
	n, ok := args[1].(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return value.NewFault("machine.sub_small: amount must be integer")
	}
	return newSame(a.Kind, uint32(int64(a.Bits)-n.Val.Num().Int64()))
}
func halMachineLT(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.lt: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.lt")
	if f != nil {
		return f
	}
	b, f := machineScalar(args[1], "machine.lt")
	if f != nil {
		return f
	}
	if a.Kind != b.Kind {
		return value.NewFault("machine.lt: machine widths differ")
	}
	if a.Bits < b.Bits {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineGT(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.gt: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.gt")
	if f != nil {
		return f
	}
	b, f := machineScalar(args[1], "machine.gt")
	if f != nil {
		return f
	}
	if a.Kind != b.Kind {
		return value.NewFault("machine.gt: machine widths differ")
	}
	if a.Bits > b.Bits {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineAddCarryGeneric(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.add_carry: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.add_carry")
	if f != nil {
		return f
	}
	b, f := machineScalar(args[1], "machine.add_carry")
	if f != nil {
		return f
	}
	if a.Kind != b.Kind {
		return value.NewFault("machine.add_carry: machine widths differ")
	}
	if uint64(a.Bits)+uint64(b.Bits) > uint64(machineMask(a.Kind)) {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineSubBorrowGeneric(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.sub_borrow: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.sub_borrow")
	if f != nil {
		return f
	}
	b, f := machineScalar(args[1], "machine.sub_borrow")
	if f != nil {
		return f
	}
	if a.Kind != b.Kind {
		return value.NewFault("machine.sub_borrow: machine widths differ")
	}
	if a.Bits < b.Bits {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineShiftRightOne(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.shr1: want 1 argument")
	}
	a, f := machineScalar(args[0], "machine.shr1")
	if f != nil {
		return f
	}
	return newSame(a.Kind, a.Bits>>1)
}
func halMachineShiftLeftOne(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.shl1: want 1 argument")
	}
	a, f := machineScalar(args[0], "machine.shl1")
	if f != nil {
		return f
	}
	return newSame(a.Kind, a.Bits<<1)
}
func halMachineLowBit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("machine.low_bit: want 1 argument")
	}
	a, f := machineScalar(args[0], "machine.low_bit")
	if f != nil {
		return f
	}
	if a.Bits&1 != 0 {
		return value.TRUE
	}
	return value.FALSE
}
func halMachineSetHighBit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.set_high_bit: want 2 arguments")
	}
	a, f := machineScalar(args[0], "machine.set_high_bit")
	if f != nil {
		return f
	}
	b, ok := args[1].(*value.Boolean)
	if !ok {
		return value.NewFault("machine.set_high_bit: value must be boolean")
	}
	mask := uint32(0x8000)
	if a.Kind == "byte" {
		mask = 0x80
	}
	v := a.Bits
	if b.Val {
		v |= mask
	} else {
		v &= ^mask
	}
	return newSame(a.Kind, v)
}

func halMachineWordAnyMask(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("machine.word_any_mask: want 2 arguments")
	}
	w, f := machineValue(args[0], "word", "machine.word_any_mask")
	if f != nil {
		return f
	}
	m, ff := exactNonNegative(args[1:2], "machine.word_any_mask", 0xffff)
	if ff != nil {
		return ff
	}
	if w&uint32(m) != 0 {
		return value.TRUE
	}
	return value.FALSE
}
