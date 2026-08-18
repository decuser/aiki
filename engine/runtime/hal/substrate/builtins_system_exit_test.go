package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestSystemExitValidatesPortableStatus(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()

	for _, code := range []int64{0, 1, 127, 255} {
		got := rt.halSystemExit([]value.Value{value.NewNumber(code, 1)}, nil)
		exit, ok := got.(*value.ProgramExitSignal)
		if !ok || exit.Code != int(code) {
			t.Fatalf("system.exit(%d) = %T %v, want ProgramExitSignal(%d)", code, got, got, code)
		}
	}

	for _, bad := range []value.Value{
		value.NewNumber(-1, 1),
		value.NewNumber(256, 1),
		value.NewNumber(1, 2),
		&value.String{Val: "1"},
	} {
		if got := rt.halSystemExit([]value.Value{bad}, nil); !value.IsFault(got) {
			t.Fatalf("system.exit(%s) = %s, want fault", bad.Inspect(), got.Inspect())
		}
	}
}
