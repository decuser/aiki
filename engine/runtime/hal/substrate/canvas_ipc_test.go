package substrate

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCanvasIPCEncodeLineBasic(t *testing.T) {
	msg := CanvasIPCMsg{Op: "dot"}
	b, err := msg.encodeLine()
	if err != nil {
		t.Fatalf("encodeLine: %v", err)
	}
	s := string(b)
	if !strings.HasSuffix(s, "\n") {
		t.Fatalf("expected newline suffix, got %q", s)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSuffix(s, "\n")), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["op"] != "dot" {
		t.Fatalf("op mismatch: %v", got["op"])
	}
	if _, ok := got["args"]; ok {
		t.Fatalf("args should be omitted when empty")
	}
	if _, ok := got["rgba"]; ok {
		t.Fatalf("rgba should be omitted when empty")
	}
	if _, ok := got["pen"]; ok {
		t.Fatalf("pen should be omitted when zero")
	}
}

func TestCanvasIPCEncodeLineFields(t *testing.T) {
	msg := CanvasIPCMsg{
		Op:   "line",
		Args: []int{1, 2, 3, 4},
		RGBA: []int{10, 20, 30, 40},
		Pen:  5,
	}
	b, err := msg.encodeLine()
	if err != nil {
		t.Fatalf("encodeLine: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSuffix(string(b), "\n")), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["op"] != "line" {
		t.Fatalf("op mismatch: %v", got["op"])
	}
	if got["pen"] != float64(5) {
		t.Fatalf("pen mismatch: %v", got["pen"])
	}
	if _, ok := got["args"]; !ok {
		t.Fatalf("args missing")
	}
	if _, ok := got["rgba"]; !ok {
		t.Fatalf("rgba missing")
	}
}
