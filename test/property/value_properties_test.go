package property

import (
	"math/big"
	"testing"
	"testing/quick"

	"aiki/engine/semantics/value"
)

// TestNumberArithmeticPreservesRat verifies arithmetic on Numbers always returns *big.Rat.
func TestNumberArithmeticPreservesRat(t *testing.T) {
	// Property: adding two numbers produces a number with *big.Rat
	add := func(a, b int64) bool {
		if b == 0 {
			b = 1 // avoid division issues in later tests
		}
		n1 := value.NewNumber(a, 1)
		n2 := value.NewNumber(b, 1)

		// Simulate addition (what evaluator does)
		result := new(big.Rat).Add(n1.Val, n2.Val)
		n3 := &value.Number{Val: result}

		return n3.Val != nil && n3.Type() == value.NumberType
	}

	if err := quick.Check(add, nil); err != nil {
		t.Error(err)
	}
}

// TestNumberDivisionPreservesRat verifies division produces rational, not float.
func TestNumberDivisionPreservesRat(t *testing.T) {
	div := func(a, b int64) bool {
		if b == 0 {
			return true // skip division by zero
		}
		n1 := value.NewNumber(a, 1)
		n2 := value.NewNumber(b, 1)

		result := new(big.Rat).Quo(n1.Val, n2.Val)
		n3 := &value.Number{Val: result}

		// Key property: result is exact rational, not float approximation
		return n3.Val != nil && n3.Val.IsInt() == (a%b == 0)
	}

	if err := quick.Check(div, nil); err != nil {
		t.Error(err)
	}
}

// TestListLengthNonNegative verifies list length is always >= 0.
func TestListLengthNonNegative(t *testing.T) {
	lengthNonNeg := func(n uint8) bool {
		// Create list with n elements
		elements := make([]value.Value, n)
		for i := range elements {
			elements[i] = value.NewNumber(int64(i), 1)
		}
		list := &value.List{Elements: elements}

		return len(list.Elements) >= 0 && len(list.Elements) == int(n)
	}

	if err := quick.Check(lengthNonNeg, nil); err != nil {
		t.Error(err)
	}
}

// TestStringInspectRoundtrip verifies strings inspect to quoted form.
func TestStringInspectRoundtrip(t *testing.T) {
	inspect := func(s string) bool {
		// Skip strings with problematic characters for this simple test
		for _, r := range s {
			if r < 32 || r > 126 {
				return true // skip non-printable
			}
		}

		str := &value.String{Val: s}
		inspected := str.Inspect()

		// Inspect should produce quoted string
		if len(inspected) < 2 {
			return false
		}
		return inspected[0] == '"' && inspected[len(inspected)-1] == '"'
	}

	if err := quick.Check(inspect, nil); err != nil {
		t.Error(err)
	}
}

// TestSymbolInspectHasColon verifies symbols inspect with leading colon.
func TestSymbolInspectHasColon(t *testing.T) {
	inspect := func(s string) bool {
		if s == "" {
			return true // skip empty
		}
		// Only valid identifier characters
		for _, r := range s {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return true // skip invalid symbol names
			}
		}

		sym := &value.Symbol{Val: s}
		inspected := sym.Inspect()

		return len(inspected) > 0 && inspected[0] == ':'
	}

	if err := quick.Check(inspect, nil); err != nil {
		t.Error(err)
	}
}
