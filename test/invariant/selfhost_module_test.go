package invariant

import (
	"strconv"
	"strings"
	"testing"
)

// TestSelfhostModuleLoading proves that representative ordinary and HAL-backed
// library modules are read, lexed, parsed, evaluated, exported, and consumed by
// the independent Aiki interpreter rather than delegated to host import.
func TestSelfhostModuleLoading(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)
	cases := []struct {
		name   string
		source string
		probe  string
	}{
		{"list", "let list = import(\"list\")\n", `list.sum([1, 2, 3, 4])`},
		{"string transitive", "let string = import(\"string\")\n", `string.trim("  hello  ")`},
		{"math ffi", "let math = import(\"math/ffi\")\n", `math.floor(7/2)`},
		{"selective import", "import(\"list\", :sum)\n", `sum([2, 3, 4])`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			direct := tc.source + "println(inspect(" + tc.probe + "))\n"
			wantOut, wantErr, wantCode, err := runAikiSource(exe, direct, root)
			if err != nil {
				t.Fatalf("run reference: %v", err)
			}
			if wantCode != 0 {
				t.Fatalf("reference exited %d\nstderr: %s", wantCode, wantErr)
			}
			source := tc.source + tc.probe + "\n"
			harness := "let bootstrap = import(\"selfhost/bootstrap\")\n" +
				"println(inspect(bootstrap.run(" + strconv.Quote(source) + ", \"module-probe.ai\")))\n"
			gotOut, gotErr, gotCode, err := runAikiSource(exe, harness, root)
			if err != nil {
				t.Fatalf("run self-host: %v", err)
			}
			if gotCode != 0 {
				t.Fatalf("self-host exited %d\nstderr: %s", gotCode, gotErr)
			}
			if strings.TrimSpace(gotOut) != strings.TrimSpace(wantOut) {
				t.Fatalf("module result differs\nreference: %q\nself-host: %q\nstderr: %s", strings.TrimSpace(wantOut), strings.TrimSpace(gotOut), gotErr)
			}
		})
	}
}
