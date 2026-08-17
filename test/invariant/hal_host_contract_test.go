package invariant

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
)

// TestHALHostOperationThreeNameCoverage makes the M1 three-name relationship
// executable. Every canonical host contract must have one compatibility
// primitive, one substrate provenance name, and at least one Aiki-facing owner.
// The Aiki source named by the descriptor must actually depend on that primitive.
func TestHALHostOperationThreeNameCoverage(t *testing.T) {
	root := distributionRoot(t)
	rt := substrate.NewGoRuntime()
	ops := rt.HostOperations()
	if len(ops) != 46 {
		t.Fatalf("expected 46 canonical host operations, got %d", len(ops))
	}

	visible := map[string]bool{}
	for _, name := range rt.BuiltinNames(value.ScopePrelude) {
		visible[name] = true
	}

	identities := map[string]string{}
	primitives := map[string]string{}
	for _, op := range ops {
		if !strings.HasPrefix(op.Identity, "HAL.") {
			t.Errorf("%s: canonical identity must begin HAL.", op.Primitive)
		}
		if op.Authority != op.Identity {
			t.Errorf("%s: M1 authority must be its exact canonical identity, got %q", op.Identity, op.Authority)
		}
		if op.SubstrateProvenance == "" || !strings.HasPrefix(op.SubstrateProvenance, "go:") {
			t.Errorf("%s: missing Go substrate provenance", op.Identity)
		}
		if len(op.Context) == 0 || op.Context[0] == "" {
			t.Errorf("%s: missing required context facet", op.Identity)
		}
		if op.Effect == "" || op.Blocking == "" || op.Lifetime == "" || op.Optionality == "" || op.ErrorContract == "" {
			t.Errorf("%s: incomplete host contract metadata", op.Identity)
		}
		if op.SemanticObservation == "" {
			t.Errorf("%s: missing Aiki-level semantic observation identity", op.Identity)
		}
		if prior, ok := identities[op.Identity]; ok {
			t.Errorf("duplicate HAL identity %s on %s and %s", op.Identity, prior, op.Primitive)
		}
		identities[op.Identity] = op.Primitive
		if prior, ok := primitives[op.Primitive]; ok {
			t.Errorf("primitive %s bound by both %s and %s", op.Primitive, prior, op.Identity)
		}
		primitives[op.Primitive] = op.Identity
		if !visible[op.Primitive] {
			t.Errorf("%s: compatibility primitive %s is not registered", op.Identity, op.Primitive)
		}
		if len(op.AikiBindings) == 0 {
			t.Errorf("%s: no Aiki programmer-facing binding recorded", op.Identity)
		}

		for _, b := range op.AikiBindings {
			assertAikiBindingUsesPrimitive(t, root, b.Name, b.Source, op.Primitive)
			authority := rt.AuthorityForSource(b.Source)
			if !authority.Allows(op.Identity) {
				t.Errorf("%s: trusted source %s lacks canonical authority %s", op.Identity, b.Source, op.Identity)
			}
			if authority.Allows(op.Primitive) {
				t.Errorf("%s: trusted source %s grants raw host binding %s instead of canonical identity", op.Identity, b.Source, op.Primitive)
			}
		}
	}
}

func assertAikiBindingUsesPrimitive(t *testing.T, root, qualifiedName, source, primitive string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(source))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%s: read Aiki binding source %s: %v", qualifiedName, source, err)
		return
	}

	name := qualifiedName
	if dot := strings.LastIndex(name, "."); dot >= 0 {
		name = name[dot+1:]
	}
	pattern := `(?ms)^let\s+` + regexp.QuoteMeta(name) + `\s*=\s*\([^)]*\)\s*\{([^}]*)\}`
	re := regexp.MustCompile(pattern)
	match := re.FindSubmatch(data)
	if match == nil {
		t.Errorf("%s: cannot find wrapper %s in %s", qualifiedName, name, source)
		return
	}
	if !strings.Contains(string(match[1]), primitive) {
		t.Errorf("%s: wrapper in %s does not depend on %s", qualifiedName, source, primitive)
	}
}
