package value

import "unicode/utf8"

// RuneLen returns the number of Unicode code points in the string without
// materializing a []rune representation.
func (s *String) RuneLen() int {
	if s == nil {
		return 0
	}
	return utf8.RuneCountInString(s.Val)
}

// RuneAt returns the Unicode code point at logical rune index i without
// materializing the whole string. Invalid UTF-8 follows Go's range/[]rune
// semantics: each invalid byte is observed as utf8.RuneError.
func (s *String) RuneAt(i int) (rune, bool) {
	if s == nil || i < 0 {
		return 0, false
	}
	at := 0
	for _, r := range s.Val {
		if at == i {
			return r, true
		}
		at++
	}
	return 0, false
}

// FirstRune returns the first Unicode code point without materializing the
// string as []rune.
func (s *String) FirstRune() (rune, bool) {
	return s.RuneAt(0)
}

// CompareRunes preserves Aiki's rune-wise string ordering without allocating
// temporary rune slices.
func (s *String) CompareRunes(other *String) int {
	if s == nil {
		if other == nil || other.Val == "" {
			return 0
		}
		return -1
	}
	if other == nil {
		if s.Val == "" {
			return 0
		}
		return 1
	}

	left := s.Val
	right := other.Val
	for len(left) > 0 && len(right) > 0 {
		lr, ls := utf8.DecodeRuneInString(left)
		rr, rs := utf8.DecodeRuneInString(right)
		switch {
		case lr < rr:
			return -1
		case lr > rr:
			return 1
		}
		left = left[ls:]
		right = right[rs:]
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
