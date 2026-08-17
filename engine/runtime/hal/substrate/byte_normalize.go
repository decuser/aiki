package substrate

import (
	"fmt"

	"aiki/engine/semantics/value"
)

// bytesFromAiki accepts either Aiki's opaque host-backed bytes value or the
// pure-Aiki @bytes shaped-list representation and returns host bytes.
func bytesFromAiki(v value.Value) ([]byte, error) {
	switch b := v.(type) {
	case *value.Bytes:
		return b.Val, nil
	case *value.List:
		if b.Shape != "bytes" {
			return nil, fmt.Errorf("expected bytes, got %s", b.Inspect())
		}
		out := make([]byte, len(b.Elements))
		for i, elem := range b.Elements {
			n, ok := elem.(*value.Number)
			if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
				return nil, fmt.Errorf("byte element %d must be an integer", i)
			}
			iv := n.Val.Num().Int64()
			if iv < 0 || iv > 255 {
				return nil, fmt.Errorf("byte element %d out of range (0-255): %d", i, iv)
			}
			out[i] = byte(iv)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected bytes, got %s", v.Type())
	}
}
