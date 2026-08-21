package value

import "testing"

func TestGetSourceLineUsesCachedSource(t *testing.T) {
	env := NewEnv()
	env.SetSource("one\ntwo\nthree\n")
	if got := env.GetSourceLine(2); got != "two" {
		t.Fatalf("line 2: got %q", got)
	}
	if got := env.GetSourceLine(4); got != "" {
		t.Fatalf("line 4: got %q", got)
	}
}

func TestIsolatedEnclosedEnvOwnsCallStack(t *testing.T) {
	parent := NewEnvWithScope(ScopePrelude)
	parent.PushFrame("parent", 1, ScopePrelude)

	child := NewIsolatedEnclosedEnv(parent)
	if got := child.StackDepth(); got != 0 {
		t.Fatalf("isolated child stack depth: got %d, want 0", got)
	}
	child.PushFrame("child", 2, ScopePrelude)
	if got := parent.StackDepth(); got != 1 {
		t.Fatalf("parent stack changed by child: got %d, want 1", got)
	}
	if got := child.StackDepth(); got != 1 {
		t.Fatalf("child stack depth: got %d, want 1", got)
	}
}

func TestCallEnvUsesLexicalBindingsAndDynamicStack(t *testing.T) {
	lexical := NewEnvWithScope(ScopePrelude)
	lexical.Set("lexical", NewNumber(7, 1))
	lexical.PushFrame("lexical-frame", 1, ScopePrelude)

	caller := NewIsolatedEnclosedEnv(lexical)
	caller.PushFrame("caller-frame", 2, ScopePrelude)

	call := NewCallEnv(lexical, caller)
	if got, ok := call.Get("lexical"); !ok || got.Inspect() != "7" {
		t.Fatalf("lexical binding unavailable: %v %v", got, ok)
	}
	if got := call.StackDepth(); got != 1 {
		t.Fatalf("call stack depth: got %d, want caller depth 1", got)
	}
	frame, ok := call.CurrentFrame()
	if !ok || frame.Name != "caller-frame" {
		t.Fatalf("call did not inherit dynamic caller frame: %#v", frame)
	}
}

func TestScopeDoesNotConferAuthority(t *testing.T) {
	env := NewEnvWithScope(ScopePrelude)
	if env.GetAuthority().Allows("_print") {
		t.Fatal("ScopePrelude must not confer raw primitive authority")
	}

	env.SetAuthority(NewAuthority("_print"))
	if !env.GetAuthority().Allows("_print") {
		t.Fatal("explicit authority grant was not retained")
	}
	if env.GetAuthority().Allows("_file_open") {
		t.Fatal("authority must grant only declared primitives")
	}
}

func TestIsolatedEnvSeparatesPreludeVocabularyFromAuthority(t *testing.T) {
	prelude := NewEnvWithScope(ScopePrelude)
	prelude.SetAuthority(NewAuthority("_print", "_file_open"))
	prelude.Set("println", &String{Val: "visible-vocabulary"})

	spawnedAuthority := NewAuthority("_print")
	isolated := NewIsolatedEnclosedEnvWithAuthority(prelude, spawnedAuthority)

	if _, ok := isolated.Get("println"); !ok {
		t.Fatal("isolated env must retain access to prelude lexical vocabulary")
	}
	if !isolated.GetAuthority().Allows("_print") {
		t.Fatal("isolated env must retain definition-bound authority")
	}
	if isolated.GetAuthority().Allows("_file_open") {
		t.Fatal("isolated env must not inherit outer prelude authority")
	}
}

func TestCallEnvBorrowedParamsPreserveBindingSemantics(t *testing.T) {
	outer := NewEnv()
	outer.Set("x", NewNumber(99, 1))
	caller := NewEnv()
	env := NewCallEnv(outer, caller)
	one := NewNumber(1, 1)
	two := NewNumber(2, 1)
	args := []Value{one, two}
	env.BindCallParams([]string{"x", "y"}, args)

	if got, ok := env.Get("x"); !ok || got != one {
		t.Fatalf("borrowed param x: got %v, %v", got, ok)
	}
	three := NewNumber(3, 1)
	if !env.Update("x", three) {
		t.Fatal("update borrowed param x failed")
	}
	if got, _ := env.Get("x"); got != three {
		t.Fatalf("updated param x: got %v want %v", got, three)
	}
	if !env.Delete("x") {
		t.Fatal("delete borrowed param x failed")
	}
	if got, ok := env.Get("x"); !ok || got == three {
		t.Fatalf("deleted param should reveal outer x: got %v, %v", got, ok)
	}
	if env.Delete("x") {
		t.Fatal("second delete of borrowed param should report false")
	}
	env.Set("x", one)
	if got, _ := env.Get("x"); got != one {
		t.Fatalf("set should restore borrowed param: got %v want %v", got, one)
	}
}

func TestResetCallEnvClearsInvocationState(t *testing.T) {
	outer1 := NewEnv()
	outer1.Set("outer", NewNumber(1, 1))
	caller := NewEnv()
	env := NewCallEnv(outer1, caller)
	args := []Value{NewNumber(2, 1)}
	env.BindCallParams([]string{"x"}, args)
	env.Set("local", NewNumber(3, 1))
	env.DefineShape(&ShapeDef{Name: "temp"})
	env.Delete("x")

	outer2 := NewEnvWithScope(ScopePrelude)
	outer2.Set("next", NewNumber(4, 1))
	env.ResetCallEnv(outer2, caller)

	if _, ok := env.Get("local"); ok {
		t.Fatal("local survived ResetCallEnv")
	}
	if _, ok := env.GetShape("temp"); ok {
		t.Fatal("shape survived ResetCallEnv")
	}
	if _, ok := env.Get("x"); ok {
		t.Fatal("parameter survived ResetCallEnv")
	}
	if got, ok := env.Get("next"); !ok || got.Inspect() != "4" {
		t.Fatalf("new lexical outer unavailable: %v %v", got, ok)
	}
	if env.GetScope() != ScopePrelude {
		t.Fatalf("scope not reset: %v", env.GetScope())
	}
}

func TestEnvColdMetadataPreservesInheritanceAndReset(t *testing.T) {
	root := NewEnv()
	root.SetFile("root.ai")
	root.SetSource("one\ntwo\n")
	root.SetExports([]string{"x"})
	root.SetPackageName("pkg")
	root.DefineShape(&ShapeDef{Name: "point"})

	child := NewEnclosedEnv(root)
	if got := child.GetFile(); got != "root.ai" {
		t.Fatalf("child file = %q, want root.ai", got)
	}
	if got := child.GetSourceLine(2); got != "two" {
		t.Fatalf("child source line = %q, want two", got)
	}
	if _, ok := child.GetShape("point"); !ok {
		t.Fatal("child must inherit root shape")
	}

	call := NewCallEnv(root, child)
	call.SetFile("call.ai")
	call.SetSource("call\nsource\n")
	call.SetExports([]string{"temp"})
	call.SetPackageName("temporary")
	call.DefineShape(&ShapeDef{Name: "temp"})
	call.ResetCallEnv(root, child)

	if got := call.GetFile(); got != "root.ai" {
		t.Fatalf("reset call file = %q, want inherited root.ai", got)
	}
	if got := call.GetSourceLine(2); got != "two" {
		t.Fatalf("reset call source line = %q, want inherited two", got)
	}
	if got := call.GetExports(); got != nil {
		t.Fatalf("reset call exports = %#v, want nil", got)
	}
	if got := call.GetPackageName(); got != "" {
		t.Fatalf("reset call package = %q, want empty", got)
	}
	if _, ok := call.GetShape("temp"); ok {
		t.Fatal("reset call must discard invocation-local shape")
	}
	if _, ok := call.GetShape("point"); !ok {
		t.Fatal("reset call must retain inherited shape lookup")
	}
}

func TestEnvCompactBindingsPromoteAtFifthLocal(t *testing.T) {
	env := NewEnv()
	env.bindings = nil

	for i, name := range []string{"a", "b", "c", "d"} {
		env.Set(name, &Number{Val: big.NewRat(int64(i+1), 1)})
		if env.bindings == nil {
			t.Fatalf("binding block missing after %s", name)
		}
		if env.bindings.spill != nil {
			t.Fatalf("binding %s promoted before fifth local", name)
		}
	}

	env.Set("e", &Number{Val: big.NewRat(5, 1)})
	if env.bindings.spill == nil {
		t.Fatal("fifth local must promote compact bindings to map")
	}

	for i, name := range []string{"a", "b", "c", "d", "e"} {
		got, ok := env.Get(name)
		if !ok {
			t.Fatalf("missing %s after promotion", name)
		}
		n, ok := got.(*Number)
		if !ok || n.Val.Cmp(big.NewRat(int64(i+1), 1)) != 0 {
			t.Fatalf("%s = %v, want %d", name, got, i+1)
		}
	}
}

func TestEnvCompactBindingsDeleteUpdateAndOuterReveal(t *testing.T) {
	outer := NewEnv()
	outer.Set("x", &String{Val: "outer"})
	inner := NewEnclosedEnv(outer)
	inner.Set("x", &String{Val: "inner"})
	inner.Set("y", &String{Val: "one"})

	if !inner.Update("y", &String{Val: "two"}) {
		t.Fatal("compact local update failed")
	}
	if got, _ := inner.Get("y"); got.(*String).Val != "two" {
		t.Fatalf("updated y = %v", got)
	}
	if !inner.Delete("x") {
		t.Fatal("compact local delete failed")
	}
	if got, ok := inner.Get("x"); !ok || got.(*String).Val != "outer" {
		t.Fatalf("delete must reveal outer x, got %v %v", got, ok)
	}
}

func TestResetCallEnvReturnsPromotedBindingsToCompactEmpty(t *testing.T) {
	root := NewEnv()
	call := NewCallEnv(root, root)
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		call.Set(name, EMPTY)
	}
	if call.bindings == nil || call.bindings.spill == nil {
		t.Fatal("setup must promote")
	}

	call.ResetCallEnv(root, root)
	if call.bindings == nil {
		t.Fatal("tail reset should retain external binding block allocation")
	}
	if call.bindings.spill != nil || call.bindings.n != 0 {
		t.Fatal("tail reset must return retained block to compact-empty state")
	}
	call.Set("z", EMPTY)
	if call.bindings.spill != nil || call.bindings.n != 1 {
		t.Fatal("post-reset first local must use compact representation")
	}
}
