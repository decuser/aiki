package value

import "testing"

func TestCompareNatural(t *testing.T) {
	tests := []struct {
		name        string
		left, right Value
		want        int
		wantOK      bool
	}{
		{"numbers", NewNumber(1, 2), NewNumber(2, 3), -1, true},
		{"strings by runes", &String{Val: "é"}, &String{Val: "猫"}, -1, true},
		{"runes", &Rune{Val: 'λ'}, &Rune{Val: '猫'}, -1, true},
		{"symbols", &Symbol{Val: "alpha"}, &Symbol{Val: "beta"}, -1, true},
		{"mixed", NewNumber(1, 1), &String{Val: "1"}, 0, false},
		{"lists", &List{}, &List{}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CompareNatural(tt.left, tt.right)
			if ok != tt.wantOK || got != tt.want {
				t.Fatalf("CompareNatural(%s, %s) = (%d, %v), want (%d, %v)", tt.left.Inspect(), tt.right.Inspect(), got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
