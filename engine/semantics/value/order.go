package value

// CompareNatural compares Aiki values that have a language-defined natural
// ordering. It returns cmp < 0 when left precedes right, cmp == 0 when equal,
// cmp > 0 when left follows right, and ok == false when the values are not
// naturally comparable. Natural ordering is defined only for values of the
// same scalar type: numbers, strings, runes, and symbols.
func CompareNatural(left, right Value) (cmp int, ok bool) {
	switch l := left.(type) {
	case *Number:
		r, same := right.(*Number)
		if !same {
			return 0, false
		}
		return l.Compare(r), true
	case *String:
		r, same := right.(*String)
		if !same {
			return 0, false
		}
		return compareRunes([]rune(l.Val), []rune(r.Val)), true
	case *Rune:
		r, same := right.(*Rune)
		if !same {
			return 0, false
		}
		switch {
		case l.Val < r.Val:
			return -1, true
		case l.Val > r.Val:
			return 1, true
		default:
			return 0, true
		}
	case *Symbol:
		r, same := right.(*Symbol)
		if !same {
			return 0, false
		}
		return compareRunes([]rune(l.Val), []rune(r.Val)), true
	default:
		return 0, false
	}
}

func compareRunes(left, right []rune) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for i := 0; i < limit; i++ {
		switch {
		case left[i] < right[i]:
			return -1
		case left[i] > right[i]:
			return 1
		}
	}
	switch {
	case len(left) < len(right):
		return -1
	case len(left) > len(right):
		return 1
	default:
		return 0
	}
}
