package repl

import "testing"

func TestDisplayLines(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  int
	}{
		{"one line", "hello", 80, 1},
		{"explicit lines", "one\ntwo\nthree", 80, 3},
		{"wrapped line", "123456789", 4, 3},
		{"blank line", "one\n\nthree", 80, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := displayLines(tc.text, tc.width); got != tc.want {
				t.Fatalf("displayLines(%q, %d) = %d, want %d", tc.text, tc.width, got, tc.want)
			}
		})
	}
}

func TestPagerCommandDefault(t *testing.T) {
	name, args := pagerCommand("")
	if name != "less" || len(args) != 1 || args[0] != "-R" {
		t.Fatalf("pagerCommand(\"\") = %q %v, want less [-R]", name, args)
	}
}

func TestPagerCommandEnvironment(t *testing.T) {
	name, args := pagerCommand("less -S -R")
	if name != "less" || len(args) != 2 || args[0] != "-S" || args[1] != "-R" {
		t.Fatalf("pagerCommand env = %q %v", name, args)
	}
}
