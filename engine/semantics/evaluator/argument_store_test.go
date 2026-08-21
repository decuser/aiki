package evaluator

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestReusableArgTargetRequiresNonEscapingNonRestFunction(t *testing.T) {
	safe := &value.Function{TailEnvReusable: true}
	if !reusableArgTarget(safe) {
		t.Fatal("closure-incapable non-rest function should permit reusable arguments")
	}
	if reusableArgTarget(&value.Function{TailEnvReusable: false}) {
		t.Fatal("closure-capable function must use durable argument storage")
	}
	if reusableArgTarget(&value.Function{TailEnvReusable: true, Rest: "rest"}) {
		t.Fatal("rest function must use durable argument storage")
	}
}

func TestArgFrameInlineAndPromotion(t *testing.T) {
	ev := New(nil, nil)
	env := value.NewEnv()
	fn := &value.Function{TailEnvReusable: true}

	small := ev.acquireArgValues(fn, argFrameInlineCapacity, env)
	if small.Frame == nil {
		t.Fatal("small eligible arity must use reusable frame")
	}
	if small.Frame.spill != nil {
		t.Fatal("small eligible arity must stay inline")
	}
	ev.releaseEvaluatedArgs(small)

	large := ev.acquireArgValues(fn, argFrameInlineCapacity+1, env)
	if large.Frame == nil {
		t.Fatal("promoted eligible arity must still own reusable frame")
	}
	if len(large.Frame.spill) != argFrameInlineCapacity+1 {
		t.Fatalf("promoted spill len = %d, want %d", len(large.Frame.spill), argFrameInlineCapacity+1)
	}
	ev.releaseEvaluatedArgs(large)
}

func TestActiveRecursiveArgFramesNeverAlias(t *testing.T) {
	ev := New(nil, nil)
	env := value.NewEnv()
	fn := &value.Function{TailEnvReusable: true}

	outer := ev.acquireArgValues(fn, 2, env)
	inner := ev.acquireArgValues(fn, 2, env)
	if outer.Frame == nil || inner.Frame == nil {
		t.Fatal("eligible recursive calls must acquire frames")
	}
	if outer.Frame == inner.Frame {
		t.Fatal("simultaneously active recursive calls must never share argument frame")
	}

	ev.releaseEvaluatedArgs(inner)
	ev.releaseEvaluatedArgs(outer)
}

func TestArgFrameReleaseClearsReferences(t *testing.T) {
	ev := New(nil, nil)
	env := value.NewEnv()
	fn := &value.Function{TailEnvReusable: true}
	args := ev.acquireArgValues(fn, 2, env)
	frame := args.Frame
	args.Values[0] = &value.String{Val: "left"}
	args.Values[1] = &value.String{Val: "right"}

	ev.releaseEvaluatedArgs(args)
	for i, v := range frame.inline {
		if v != nil {
			t.Fatalf("released inline slot %d retained value %v", i, v)
		}
	}
}

func TestArgFrameMeasuredPromotionBoundaryIsThree(t *testing.T) {
	if argFrameInlineCapacity != 2 {
		t.Fatalf("inline argument capacity = %d, want measured capacity 2", argFrameInlineCapacity)
	}
}

func TestArgFrameReleaseClearsOnlyUsedCompactSlots(t *testing.T) {
	ev := New(nil, nil)
	env := value.NewEnv()
	fn := &value.Function{TailEnvReusable: true}

	args := ev.acquireArgValues(fn, 1, env)
	frame := args.Frame
	if frame == nil {
		t.Fatal("eligible one-argument call must acquire reusable frame")
	}

	args.Values[0] = &value.String{Val: "used"}
	// This slot is deliberately outside the active view. It is not semantically
	// owned by this call and therefore must not be swept merely because the
	// frame has compact capacity for it.
	frame.inline[1] = &value.String{Val: "sentinel"}

	ev.releaseEvaluatedArgs(args)

	if frame.inline[0] != nil {
		t.Fatal("used compact slot retained value after release")
	}
	if frame.inline[1] == nil {
		t.Fatal("unused compact slot was unnecessarily cleared")
	}

	// Restore test-created sentinel before returning the frame to normal use.
	frame.inline[1] = nil
}
