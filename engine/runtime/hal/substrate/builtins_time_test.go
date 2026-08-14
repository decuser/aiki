package substrate

import (
	"testing"
	"time"

	"aiki/engine/semantics/value"
)

func TestHalAfterProducesReceiveOnlyEvent(t *testing.T) {
	v := halAfter([]value.Value{value.NewNumber(0, 1)}, nil)
	ch, ok := v.(*value.Channel)
	if !ok {
		t.Fatalf("after returned %T, want channel", v)
	}
	if ch.CanSend() {
		t.Fatal("timer channel must be receive-only")
	}
	if cap(ch.C) != 1 {
		t.Fatalf("timer channel capacity: got %d, want 1", cap(ch.C))
	}

	select {
	case got := <-ch.C:
		if got != value.TRUE {
			t.Fatalf("timer payload: got %s, want true", got.Inspect())
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("zero-duration timer did not fire")
	}
}

func TestHalAfterRejectsNegativeDuration(t *testing.T) {
	v := halAfter([]value.Value{value.NewNumber(-1, 1)}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Fatalf("after(-1): got %T, want fault", v)
	}
}

func TestHalSendRejectsReceiveOnlyChannel(t *testing.T) {
	ch := value.NewEventChannel()
	v := halSend([]value.Value{ch, value.TRUE}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Fatalf("send(receive-only): got %T, want fault", v)
	}
}

func TestOrdinaryChannelRemainsSendableAndUnbuffered(t *testing.T) {
	ch := value.NewChannel()
	if !ch.CanSend() {
		t.Fatal("ordinary channel should be sendable")
	}
	if cap(ch.C) != 0 {
		t.Fatalf("ordinary channel capacity: got %d, want 0", cap(ch.C))
	}
}
