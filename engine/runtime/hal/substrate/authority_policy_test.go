package substrate

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func TestAuthorityPolicyIsExplicitAndNarrow(t *testing.T) {
	rt := NewGoRuntime()

	fileAuth := rt.AuthorityForSource("/tmp/project/lib/file/file.ai")
	if !fileAuth.Allows("HAL.file.open") || !fileAuth.Allows("HAL.file.read_text") {
		t.Fatal("file module missing declared file authority")
	}
	if fileAuth.Allows("HAL.canvas.open") || fileAuth.Allows("_test_run") {
		t.Fatal("file module acquired unrelated raw authority")
	}

	unlisted := rt.AuthorityForSource("/tmp/project/lib/example/example.ai")
	if unlisted.Allows("HAL.io.print") || unlisted.Allows("HAL.file.open") {
		t.Fatal("lib path alone must not confer raw authority")
	}
}

func TestPreludeAuthorityIsDeclaredNotUniversal(t *testing.T) {
	rt := NewGoRuntime()
	auth := rt.AuthorityForSource("engine/runtime/prelude/prelude.ai")
	if !auth.Allows("HAL.io.print") || !auth.Allows("_length") {
		t.Fatal("prelude missing declared primitive authority")
	}
	if auth.Allows("HAL.file.open") || auth.Allows("HAL.canvas.open") {
		t.Fatal("prelude must not receive unrelated library host authority")
	}
}

func TestAuthorityPolicyMatchesTrustedSourceDependencies(t *testing.T) {
	rt := NewGoRuntime()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate authority policy test source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(here), "../../../.."))
	ident := regexp.MustCompile(`\b_[A-Za-z][A-Za-z0-9_]*\b`)

	for source, declared := range hal.AuthorityPolicy() {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(source)))
		if err != nil {
			t.Fatalf("%s: reading trusted source: %v", source, err)
		}

		actual := map[string]bool{}
		for _, name := range ident.FindAllString(string(data), -1) {
			if _, registered := rt.lookupBuiltin(name); registered {
				actual[name] = true
			}
		}
		want := map[string]bool{}
		for _, name := range declared {
			want[name] = true
		}

		for name := range actual {
			if !want[name] {
				t.Errorf("%s: raw dependency %s is not declared in authority policy", source, name)
			}
		}
		for name := range want {
			if !actual[name] {
				t.Errorf("%s: authority policy grants unused raw primitive %s", source, name)
			}
		}
	}
}

func TestHostBindingAuthorizationUsesCanonicalHALIdentity(t *testing.T) {
	rt := NewGoRuntime()
	canonical := value.NewAuthority("HAL.file.open")
	if !rt.HasBuiltin("_file_open", canonical) {
		t.Fatal("HAL.file.open authority must authorize the _file_open substrate binding")
	}
	raw := value.NewAuthority("_file_open")
	if rt.HasBuiltin("_file_open", raw) {
		t.Fatal("raw _file_open name must not itself grant canonical host authority")
	}
}
