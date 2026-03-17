# Property Tests

Random inputs, invariants must hold. Uses Go's `testing/quick`.

## value_properties_test.go

- Number arithmetic always produces `*big.Rat`
- Division preserves exactness
- List length is always non-negative
- String/Symbol inspect format is correct
