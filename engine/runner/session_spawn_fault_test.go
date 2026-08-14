package runner

import (
	"strings"
	"testing"
	"time"
)

func evalWithTimeout(t *testing.T, s *Session, source string) string {
	t.Helper()
	done := make(chan string, 1)
	go func() {
		v := s.Eval(source)
		done <- v.Inspect()
	}()
	select {
	case got := <-done:
		return got
	case <-time.After(2 * time.Second):
		t.Fatal("evaluation blocked after spawned fault")
		return ""
	}
}

func TestSpawnFaultInterruptsRecv(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	got := evalWithTimeout(t, s, `
let ch = channel()
let worker = (out) {
    let x = 1 / 0
    send(out, x)
}
spawn(worker, ch)
recv(ch)
`)
	if !strings.Contains(got, "division by zero") {
		t.Fatalf("expected spawned division fault, got %s", got)
	}
}

func TestSpawnFaultInterruptsSelect(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	got := evalWithTimeout(t, s, `
let ch = channel()
let worker = (out) {
    let x = 1 / 0
    send(out, x)
}
spawn(worker, ch)
select {
    recv(ch) { :done }
}
`)
	if !strings.Contains(got, "division by zero") {
		t.Fatalf("expected spawned division fault, got %s", got)
	}
}
