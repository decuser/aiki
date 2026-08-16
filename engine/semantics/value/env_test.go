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
