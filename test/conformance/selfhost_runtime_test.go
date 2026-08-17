package conformance

import (
	"strings"
	"testing"
)

// TestSelfhostRuntimeEnvironment proves the Phase-III runtime environment is
// implemented with ordinary Aiki closure state while preserving lexical lookup,
// shadowing, assignment through enclosing scopes, and shared capture of later
// bindings.
func TestSelfhostRuntimeEnvironment(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)

	harness := `let rt = import("./selfhost/runtime")
let root = rt.env_new([])
println(rt.env_has(root, "x"))
println(rt.env_define(root, "x", 1))
println(rt.env_get(root, "x"))
let child = rt.env_enclose(root)
println(rt.env_get(child, "x"))
println(rt.env_define(child, "x", 2))
println(rt.env_get(child, "x"))
println(rt.env_get(root, "x"))
println(rt.env_assign(child, "x", 3))
println(rt.env_get(child, "x"))
println(rt.env_assign(child, "y", 4))
let capture = rt.env_enclose(root)
let read_later = () { rt.env_get(capture, "later") }
rt.env_define(capture, "later", 9)
println(read_later())
`

	stdout, stderr, exitCode, err := runAikiSource(exe, harness, root)
	if err != nil {
		t.Fatalf("run self-host runtime environment probe: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("self-host runtime environment probe exited %d\nstderr: %s", exitCode, stderr)
	}

	want := strings.Join([]string{
		"false",
		"1",
		"1",
		"1",
		"2",
		"2",
		"1",
		"true",
		"3",
		"false",
		"9",
	}, "\n")
	if got := strings.TrimSpace(stdout); got != want {
		t.Fatalf("self-host runtime environment disagrees\nwant:\n%s\n\ngot:\n%s\n\nstderr:\n%s", want, got, stderr)
	}
}
