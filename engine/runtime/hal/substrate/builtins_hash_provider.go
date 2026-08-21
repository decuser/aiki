package substrate

import (
	"math/big"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

const hashBucketCount = 64

func hashStringProvider(s string) *big.Int {
	h := new(big.Int)
	thirtyOne := big.NewInt(31)
	for _, r := range s {
		h.Mul(h, thirtyOne)
		h.Add(h, big.NewInt(int64(r)))
	}
	return h
}

func hashCodeProvider(v value.Value) *big.Int {
	switch x := v.(type) {
	case *value.Number:
		return hashStringProvider(x.RatString())
	case *value.Boolean:
		if x.Val {
			return big.NewInt(1)
		}
		return new(big.Int)
	case *value.Rune:
		return big.NewInt(int64(x.Val))
	case *value.String:
		return hashStringProvider(x.Val)
	case *value.Symbol:
		return hashStringProvider(":" + x.Val)
	case *value.List:
		shape := x.Shape
		if shape == "" {
			shape = "list"
		}
		h := hashStringProvider(":" + shape)
		thirtyOne := big.NewInt(31)
		for _, elem := range x.Elements {
			h.Mul(h, thirtyOne)
			h.Add(h, hashCodeProvider(elem))
		}
		return h
	default:
		return new(big.Int)
	}
}

func hashNumber(n *big.Int) value.Value {
	return value.NewNumberFromBigInt(new(big.Int).Set(n))
}

func hashBuckets(v value.Value, op string) (*value.List, *value.Fault) {
	h, ok := v.(*value.List)
	if !ok {
		return nil, value.NewFault("%s: expected hash map", op)
	}
	if len(h.Elements) != hashBucketCount {
		return nil, value.NewFault("%s: expected %d hash buckets, got %d", op, hashBucketCount, len(h.Elements))
	}
	return h, nil
}

func hashBucketIndex(key value.Value) int {
	idx := new(big.Int).Mod(hashCodeProvider(key), big.NewInt(hashBucketCount))
	return int(idx.Int64())
}

func hashBucketAt(h *value.List, idx int, op string) (*value.List, *value.Fault) {
	bucket, ok := h.Elements[idx].(*value.List)
	if !ok {
		return nil, value.NewFault("%s: bucket %d is not a list", op, idx)
	}
	return bucket, nil
}

func hashPair(v value.Value, op string) (*value.List, *value.Fault) {
	pair, ok := v.(*value.List)
	if !ok || len(pair.Elements) < 2 {
		return nil, value.NewFault("%s: invalid hash entry", op)
	}
	return pair, nil
}

func halHashCode(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("hash_code: want 1 argument, got %d", len(args))
	}
	return hashNumber(hashCodeProvider(args[0]))
}

func halHashNew(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("hash_new: want 0 arguments, got %d", len(args))
	}
	buckets := make([]value.Value, hashBucketCount)
	for i := range buckets {
		buckets[i] = &value.List{Elements: []value.Value{}}
	}
	return &value.List{Elements: buckets}
}

func halHashGet(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("hash_get: want 2 arguments, got %d", len(args))
	}
	h, fault := hashBuckets(args[0], "hash_get")
	if fault != nil {
		return fault
	}
	idx := hashBucketIndex(args[1])
	bucket, fault := hashBucketAt(h, idx, "hash_get")
	if fault != nil {
		return fault
	}
	for _, entry := range bucket.Elements {
		pair, fault := hashPair(entry, "hash_get")
		if fault != nil {
			return fault
		}
		if valuesEqual(pair.Elements[0], args[1]) {
			return pair.Elements[1]
		}
	}
	return value.NewShapedError("key", "key not found")
}

func halHashPut(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("hash_put: want 3 arguments, got %d", len(args))
	}
	h, fault := hashBuckets(args[0], "hash_put")
	if fault != nil {
		return fault
	}
	idx := hashBucketIndex(args[1])
	bucket, fault := hashBucketAt(h, idx, "hash_put")
	if fault != nil {
		return fault
	}
	newBucket := make([]value.Value, 0, len(bucket.Elements)+1)
	found := false
	for _, entry := range bucket.Elements {
		pair, fault := hashPair(entry, "hash_put")
		if fault != nil {
			return fault
		}
		if valuesEqual(pair.Elements[0], args[1]) {
			newBucket = append(newBucket, &value.List{Elements: []value.Value{args[1], args[2]}})
			found = true
		} else {
			newBucket = append(newBucket, entry)
		}
	}
	if !found {
		newBucket = append(newBucket, &value.List{Elements: []value.Value{args[1], args[2]}})
	}
	newHash := append([]value.Value(nil), h.Elements...)
	newHash[idx] = &value.List{Elements: newBucket}
	return &value.List{Elements: newHash}
}

func halHashHas(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("hash_has: want 2 arguments, got %d", len(args))
	}
	h, fault := hashBuckets(args[0], "hash_has")
	if fault != nil {
		return fault
	}
	idx := hashBucketIndex(args[1])
	bucket, fault := hashBucketAt(h, idx, "hash_has")
	if fault != nil {
		return fault
	}
	for _, entry := range bucket.Elements {
		pair, fault := hashPair(entry, "hash_has")
		if fault != nil {
			return fault
		}
		if valuesEqual(pair.Elements[0], args[1]) {
			return value.TRUE
		}
	}
	return value.FALSE
}

func halHashDel(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("hash_del: want 2 arguments, got %d", len(args))
	}
	h, fault := hashBuckets(args[0], "hash_del")
	if fault != nil {
		return fault
	}
	idx := hashBucketIndex(args[1])
	bucket, fault := hashBucketAt(h, idx, "hash_del")
	if fault != nil {
		return fault
	}
	newBucket := make([]value.Value, 0, len(bucket.Elements))
	for _, entry := range bucket.Elements {
		pair, fault := hashPair(entry, "hash_del")
		if fault != nil {
			return fault
		}
		if !valuesEqual(pair.Elements[0], args[1]) {
			newBucket = append(newBucket, entry)
		}
	}
	newHash := append([]value.Value(nil), h.Elements...)
	newHash[idx] = &value.List{Elements: newBucket}
	return &value.List{Elements: newHash}
}

func halHashKeys(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("hash_keys: want 1 argument, got %d", len(args))
	}
	h, fault := hashBuckets(args[0], "hash_keys")
	if fault != nil {
		return fault
	}
	out := []value.Value{}
	for i := 0; i < hashBucketCount; i++ {
		bucket, fault := hashBucketAt(h, i, "hash_keys")
		if fault != nil {
			return fault
		}
		for _, entry := range bucket.Elements {
			pair, fault := hashPair(entry, "hash_keys")
			if fault != nil {
				return fault
			}
			out = append(out, pair.Elements[0])
		}
	}
	return &value.List{Elements: out}
}

func halHashValues(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("hash_values: want 1 argument, got %d", len(args))
	}
	h, fault := hashBuckets(args[0], "hash_values")
	if fault != nil {
		return fault
	}
	out := []value.Value{}
	for i := 0; i < hashBucketCount; i++ {
		bucket, fault := hashBucketAt(h, i, "hash_values")
		if fault != nil {
			return fault
		}
		for _, entry := range bucket.Elements {
			pair, fault := hashPair(entry, "hash_values")
			if fault != nil {
				return fault
			}
			out = append(out, pair.Elements[1])
		}
	}
	return &value.List{Elements: out}
}
