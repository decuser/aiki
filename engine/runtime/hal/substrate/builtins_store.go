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
