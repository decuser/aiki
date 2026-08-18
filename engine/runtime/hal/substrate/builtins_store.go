package substrate

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func storeInteger(v value.Value, what string) (int64, *value.Fault) {
	n, ok := v.(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return 0, value.NewFault("%s must be an integer", what)
	}
	return n.Val.Num().Int64(), nil
}

func halStoreNew(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("store.new: want 1 argument, got %d", len(args))
	}
	size64, fault := storeInteger(args[0], "store.new: size")
	if fault != nil {
		return fault
	}
	if size64 < 0 {
		return value.NewFault("store.new: size must be non-negative")
	}
	maxInt := int64(^uint(0) >> 1)
	if size64 > maxInt {
		return value.NewFault("store.new: size is too large")
	}
	cells := make([]value.Value, int(size64))
	zero := value.NewNumber(0, 1)
	for i := range cells {
		cells[i] = zero
	}
	return &value.Store{Cells: cells}
}

func halStoreGet(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("store.get: want 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.Store)
	if !ok {
		return value.NewFault("store.get: expected store, got %s", args[0].Type())
	}
	idx64, fault := storeInteger(args[1], "store.get: index")
	if fault != nil {
		return fault
	}
	length := s.StoreLen()
	if idx64 < 0 || idx64 >= int64(length) {
		return value.NewFault("store.get: index %d out of bounds (length %d)", idx64, length)
	}
	v := s.StoreGet(int(idx64))
	semanticHit(ctx, engine.SemanticStoreRead)
	return v
}

func halStoreSet(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("store.set: want 3 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.Store)
	if !ok {
		return value.NewFault("store.set: expected store, got %s", args[0].Type())
	}
	idx64, fault := storeInteger(args[1], "store.set: index")
	if fault != nil {
		return fault
	}
	length := s.StoreLen()
	if idx64 < 0 || idx64 >= int64(length) {
		return value.NewFault("store.set: index %d out of bounds (length %d)", idx64, length)
	}
	s.StoreSet(int(idx64), args[2])
	semanticHit(ctx, engine.SemanticStoreWrite)
	return value.EMPTY
}

func halStoreLength(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("store.length: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.Store)
	if !ok {
		return value.NewFault("store.length: expected store, got %s", args[0].Type())
	}
	return value.NewNumber(int64(s.StoreLen()), 1)
}

func halStoreSnapshot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 || len(args) > 2 {
		return value.NewFault("store.snapshot: want 1 or 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.Store)
	if !ok {
		return value.NewFault("store.snapshot: expected store, got %s", args[0].Type())
	}
	count := s.StoreLen()
	if len(args) == 2 {
		count64, fault := storeInteger(args[1], "store.snapshot: count")
		if fault != nil {
			return fault
		}
		if count64 < 0 || count64 > int64(count) {
			return value.NewFault("store.snapshot: count %d out of bounds (length %d)", count64, count)
		}
		count = int(count64)
	}
	return &value.List{Elements: s.StoreSnapshot(count)}
}

func halStoreDigitsToText(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 || len(args) > 2 {
		return value.NewFault("store.digits_to_text: want 1 or 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.Store)
	if !ok {
		return value.NewFault("store.digits_to_text: expected store, got %s", args[0].Type())
	}
	count := s.StoreLen()
	if len(args) == 2 {
		count64, fault := storeInteger(args[1], "store.digits_to_text: count")
		if fault != nil {
			return fault
		}
		if count64 < 0 || count64 > int64(count) {
			return value.NewFault("store.digits_to_text: count %d out of bounds (length %d)", count64, count)
		}
		count = int(count64)
	}

	out := make([]byte, count)
	for i := 0; i < count; i++ {
		n, fault := storeInteger(s.StoreGet(i), "store.digits_to_text: cell")
		if fault != nil {
			return fault
		}
		if n < 0 || n > 9 {
			return value.NewFault("store.digits_to_text: cell %d is not a decimal digit: %d", i, n)
		}
		out[i] = byte('0' + n)
	}
	return &value.String{Val: string(out)}
}

func halStoreChecksum(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 || len(args) > 2 {
		return value.NewFault("store.checksum: want 1 or 2 arguments, got %d", len(args))
	}
	s, ok := args[0].(*value.Store)
	if !ok {
		return value.NewFault("store.checksum: expected store, got %s", args[0].Type())
	}
	count := s.StoreLen()
	if len(args) == 2 {
		count64, fault := storeInteger(args[1], "store.checksum: count")
		if fault != nil {
			return fault
		}
		if count64 < 0 || count64 > int64(count) {
			return value.NewFault("store.checksum: count %d out of bounds (length %d)", count64, count)
		}
		count = int(count64)
	}

	const modulus int64 = 1000000007
	var checksum int64 = 17
	for i := 0; i < count; i++ {
		cell, fault := storeInteger(s.StoreGet(i), "store.checksum: cell")
		if fault != nil {
			return fault
		}
		checksum = ((checksum * 131) + cell + int64(i)) % modulus
	}
	return value.NewNumber(checksum, 1)
}
