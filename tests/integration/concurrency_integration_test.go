package integration_test

import (
	"aiki/tests/testutil"
	"testing"

	"aiki/reference/semantics/value"
)

func TestChannel(t *testing.T) {
	result := testutil.EvalPrelude(`type(channel())`)
	if result.Inspect() != ":channel" {
		t.Errorf("got %s, want :channel", result.Inspect())
	}
}

func TestSendRecv(t *testing.T) {
	// FIXED: Spawn a thread to send, so the main thread can receive.
	result := testutil.EvalPrelude(`
let ch = channel()
spawn((c) {
	send(c, 42)
}, ch)
recv(ch)
`)
	testutil.TestNumberValue(t, result, "42")
}

func TestSpawnReturnsTrue(t *testing.T) {
	result := testutil.EvalPrelude(`
let ch = channel()
spawn((ch) {
	send(ch, 1)
}, ch)
`)
	b, ok := result.(*value.Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %T (%v)", result, result)
	}
	if !b.Value {
		t.Errorf("spawn should return true")
	}
}

func TestSpawnNonFunction(t *testing.T) {
	result := testutil.EvalPrelude(`spawn(42)`)
	_, ok := result.(*value.Error)
	if !ok {
		t.Errorf("expected error for spawn(42), got %T", result)
	}
}

func TestChannelMultipleValues(t *testing.T) {
	result := testutil.EvalPrelude(`
let ch = channel()

# Spawn a thread to send the values
spawn((c) {
	send(c, 1)
	send(c, 2)
	send(c, 3)
}, ch)

# Main thread receives them
let a = recv(ch)
let b = recv(ch)
let c = recv(ch)
a + b + c
`)
	testutil.TestNumberValue(t, result, "6")
}
