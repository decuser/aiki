package runner

import (
	"os"
	"testing"
	"time"
)

func TestSessionSelectReceivesReadyChannel(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	v := s.Eval(`
let a = channel()
let b = channel()
let worker = (ch) { send(ch, 42) }
spawn(worker, b)
select {
    let x = recv(a) { x + 1000 }
    let x = recv(b) { x }
}
`)
	if got := v.Inspect(); got != "42" {
		t.Fatalf("select result: got %s, want 42", got)
	}
}

func TestSessionSelectDefaultIsNonblocking(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	v := s.Eval(`
let ch = channel()
select {
    let x = recv(ch) { x }
    default { :idle }
}
`)
	if got := v.Inspect(); got != ":idle" {
		t.Fatalf("select default result: got %s, want :idle", got)
	}
}

func TestSessionSelectEvaluatesChannelExpressionsOnce(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	v := s.Eval(`
let count = 0
let ch = channel()
let source = () {
    count = count + 1
    return ch
}
select {
    recv(source()) { :received }
    default { count }
}
`)
	if got := v.Inspect(); got != "1" {
		t.Fatalf("channel expression evaluation count: got %s, want 1", got)
	}
}

func TestSessionSelectBlocksUntilReceive(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	done := make(chan string, 1)
	go func() {
		v := s.Eval(`
let ch = channel()
let worker = (out) {
    send(out, 7)
}
spawn(worker, ch)
select {
    let x = recv(ch) { x }
}
`)
		done <- v.Inspect()
	}()

	select {
	case got := <-done:
		if got != "7" {
			t.Fatalf("select result: got %s, want 7", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("select did not receive ready channel")
	}
}

func TestSessionSelectWithTimeAfter(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("chdir repo root: %v", err)
	}

	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	v := s.Eval(`
let time = import("time")
let events = channel()
select {
    recv(events) { :event }
    recv(time.after(0)) { :timeout }
}
`)
	if got := v.Inspect(); got != ":timeout" {
		t.Fatalf("select timer result: got %s, want :timeout", got)
	}
}
