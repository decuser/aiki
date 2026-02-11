package tests

import (
	"testing"
	"time"

	"aiki/lang/eval"
	"aiki/lang/value"
)

func TestChannel(t *testing.T) {
	env := value.NewEnv(nil)
	result := eval.Run("channel()", env)

	_, ok := result.(*value.Channel)
	if !ok {
		t.Fatalf("expected Channel, got %T: %v", result, result)
	}
}

func TestSendRecv(t *testing.T) {
	env := value.NewEnv(nil)

	// This will block forever without spawn, so we use goroutines directly for testing
	input := `
let ch = channel()
spawn(() {
	send(ch, 42)
})
recv(ch)
`
	done := make(chan value.Value, 1)
	go func() {
		done <- eval.Run(input, env)
	}()

	select {
	case result := <-done:
		num, ok := result.(*value.Number)
		if !ok {
			t.Fatalf("expected Number, got %T: %v", result, result)
		}
		if num.Inspect() != "42" {
			t.Errorf("got %s, want 42", num.Inspect())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: send/recv blocked")
	}
}

func TestSpawnReturnsTrue(t *testing.T) {
	env := value.NewEnv(nil)
	input := `spawn(() { let x = 1 })`

	result := eval.Run(input, env)

	b, ok := result.(*value.Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %T: %v", result, result)
	}
	if !b.Value {
		t.Error("spawn should return true")
	}
}

func TestSpawnNonFunction(t *testing.T) {
	env := value.NewEnv(nil)
	result := eval.Run("spawn(42)", env)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected Error, got %T: %v", result, result)
	}
	if err.Message != "spawn: argument must be a function" {
		t.Errorf("wrong error: %s", err.Message)
	}
}

func TestChannelMultipleValues(t *testing.T) {
	env := value.NewEnv(nil)

	input := `
let ch = channel()
spawn(() {
	send(ch, 1)
	send(ch, 2)
	send(ch, 3)
})
let a = recv(ch)
let b = recv(ch)
let c = recv(ch)
a + b + c
`
	done := make(chan value.Value, 1)
	go func() {
		done <- eval.Run(input, env)
	}()

	select {
	case result := <-done:
		num, ok := result.(*value.Number)
		if !ok {
			t.Fatalf("expected Number, got %T: %v", result, result)
		}
		if num.Inspect() != "6" {
			t.Errorf("got %s, want 6", num.Inspect())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
