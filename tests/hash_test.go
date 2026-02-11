package tests

import (
	"testing"

	"aiki/lang/eval"
	"aiki/lang/value"
)

func TestHashNew(t *testing.T) {
	env := setupEnv()
	result := eval.Run("len(hash_new())", env)
	testNumberValue(t, result, "64")
}

func TestHashPutGet(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "name", "Mochi")
hash_get(h, "name")
`
	result := eval.Run(input, env)
	str, ok := result.(*value.String)
	if !ok {
		if list, isList := result.(*value.List); isList && list.Shape == "error" {
			t.Fatalf("got error: %v", list.Elements[1])
		}
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != "Mochi" {
		t.Errorf("got %s, want Mochi", str.Value)
	}
}

func TestHashMultipleKeys(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "a", 1)
let h = hash_put(h, "b", 2)
let h = hash_put(h, "c", 3)
len(hash_keys(h))
`
	result := eval.Run(input, env)
	testNumberValue(t, result, "3")
}

func TestHashHas(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "key", "value")
hash_has(h, "key")
`
	result := eval.Run(input, env)
	testBooleanValue(t, result, true)
}

func TestHashHasMissing(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
hash_has(h, "missing")
`
	result := eval.Run(input, env)
	testBooleanValue(t, result, false)
}

func TestHashDel(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "a", 1)
let h = hash_put(h, "b", 2)
let h = hash_del(h, "a")
hash_has(h, "a")
`
	result := eval.Run(input, env)
	testBooleanValue(t, result, false)
}

func TestHashDelPreserves(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "a", 1)
let h = hash_put(h, "b", 2)
let h = hash_del(h, "a")
hash_has(h, "b")
`
	result := eval.Run(input, env)
	testBooleanValue(t, result, true)
}

func TestHashUpdate(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "x", 1)
let h = hash_put(h, "x", 2)
hash_get(h, "x")
`
	result := eval.Run(input, env)
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("expected Number, got %T", result)
	}
	if num.Inspect() != "2" {
		t.Errorf("got %s, want 2", num.Inspect())
	}
}

func TestHashValues(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, "a", 10)
let h = hash_put(h, "b", 20)
sum(hash_values(h))
`
	result := eval.Run(input, env)
	testNumberValue(t, result, "30")
}

func TestHashNumberKeys(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, 1, "one")
let h = hash_put(h, 2, "two")
hash_get(h, 1)
`
	result := eval.Run(input, env)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != "one" {
		t.Errorf("got %s, want one", str.Value)
	}
}

func TestHashSymbolKeys(t *testing.T) {
	env := setupEnv()
	input := `
let h = hash_new()
let h = hash_put(h, :name, "Mochi")
hash_get(h, :name)
`
	result := eval.Run(input, env)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != "Mochi" {
		t.Errorf("got %s, want Mochi", str.Value)
	}
}

// Helper to set up environment with prelude
func setupEnv() *value.Env {
	env := value.NewEnv(nil)
	// Load prelude shapes
	eval.Run("let @ok [value]", env)
	eval.Run("let @error [reason]", env)
	// Load hash functions
	prelude := `
let _HASH_SIZE = 64

let hash_new = () {
    let buckets = []
    let i = 0
    while i < _HASH_SIZE {
        buckets = append(buckets, [])
        i = i + 1
    }
    return buckets
}

let _bucket_index = (key) {
    return _hash_code(key) % _HASH_SIZE
}

let hash_get = (h, key) {
    let idx = _bucket_index(key)
    let bucket = nth(h, idx)
    let i = 0
    while i < len(bucket) {
        let pair = nth(bucket, i)
        if equal(nth(pair, 0), key) {
            return nth(pair, 1)
        }
        i = i + 1
    }
    return [@error, "key not found"]
}

let hash_put = (h, key, val) {
    let idx = _bucket_index(key)
    let bucket = nth(h, idx)
    let new_bucket = []
    let found = false
    let i = 0
    while i < len(bucket) {
        let pair = nth(bucket, i)
        if equal(nth(pair, 0), key) {
            new_bucket = append(new_bucket, [key, val])
            found = true
        } else {
            new_bucket = append(new_bucket, pair)
        }
        i = i + 1
    }
    if not found {
        new_bucket = append(new_bucket, [key, val])
    }
    let new_h = []
    let j = 0
    while j < _HASH_SIZE {
        if j == idx {
            new_h = append(new_h, new_bucket)
        } else {
            new_h = append(new_h, nth(h, j))
        }
        j = j + 1
    }
    return new_h
}

let hash_has = (h, key) {
    let result = hash_get(h, key)
    return shape(result) != :error
}

let hash_del = (h, key) {
    let idx = _bucket_index(key)
    let bucket = nth(h, idx)
    let new_bucket = []
    let i = 0
    while i < len(bucket) {
        let pair = nth(bucket, i)
        if not equal(nth(pair, 0), key) {
            new_bucket = append(new_bucket, pair)
        }
        i = i + 1
    }
    let new_h = []
    let j = 0
    while j < _HASH_SIZE {
        if j == idx {
            new_h = append(new_h, new_bucket)
        } else {
            new_h = append(new_h, nth(h, j))
        }
        j = j + 1
    }
    return new_h
}

let hash_keys = (h) {
    let keys = []
    let i = 0
    while i < _HASH_SIZE {
        let bucket = nth(h, i)
        let j = 0
        while j < len(bucket) {
            let pair = nth(bucket, j)
            keys = append(keys, nth(pair, 0))
            j = j + 1
        }
        i = i + 1
    }
    return keys
}

let hash_values = (h) {
    let vals = []
    let i = 0
    while i < _HASH_SIZE {
        let bucket = nth(h, i)
        let j = 0
        while j < len(bucket) {
            let pair = nth(bucket, j)
            vals = append(vals, nth(pair, 1))
            j = j + 1
        }
        i = i + 1
    }
    return vals
}

let sum = (list) {
    let result = 0
    let i = 0
    while i < len(list) {
        result = result + nth(list, i)
        i = i + 1
    }
    return result
}
`
	eval.Run(prelude, env)
	return env
}
