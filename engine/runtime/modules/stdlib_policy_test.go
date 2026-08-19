package modules

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"aiki/engine/runtime/help"
	"aiki/engine/runtime/primitives"
	"aiki/engine/syntax/grammar"
)

func TestStdlibModulePoliciesWellFormed(t *testing.T) {
	validRoles := map[StdlibSemanticRole]bool{
		RolePortable: true, RoleRuntimeCapability: true, RoleHostCapability: true,
		RoleInterop: true, RoleInternal: true,
	}
	validRealizations := map[StdlibRealization]bool{
		RealizationNative: true, RealizationFFI: true,
		RealizationIntrinsic: true, RealizationMixed: true,
	}
	seen := make(map[string]bool)
	for _, policy := range StdlibModulePolicies() {
		if policy.Module == "" {
			t.Fatal("stdlib module policy has empty module name")
		}
		if seen[policy.Module] {
			t.Fatalf("duplicate stdlib module policy for %q", policy.Module)
		}
		seen[policy.Module] = true
		if !validRoles[policy.Role] {
			t.Fatalf("stdlib module %q has invalid semantic role %q", policy.Module, policy.Role)
		}
		if !validRealizations[policy.Realization] {
			t.Fatalf("stdlib module %q has invalid realization %q", policy.Module, policy.Realization)
		}
		if policy.SemanticAuthority != "" &&
			(policy.Role != RolePortable || policy.Realization != RealizationFFI) {
			t.Fatalf("module %q declares semantic authority %q but is %s/%s",
				policy.Module, policy.SemanticAuthority, policy.Role, policy.Realization)
		}
	}
}

func TestStdlibModulePoliciesComplete(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	packageRE := regexp.MustCompile(`(?m)^package\s+"([^"]+)"`)
	var modules []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ai") || strings.HasSuffix(entry.Name(), "_test.ai") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		match := packageRE.FindSubmatch(data)
		if len(match) == 2 {
			modules = append(modules, string(match[1]))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	declared := make(map[string]bool)
	for _, policy := range StdlibModulePolicies() {
		declared[policy.Module] = true
	}
	for _, module := range modules {
		if !declared[module] {
			t.Errorf("shipped module %q has no stdlib module policy", module)
		}
		delete(declared, module)
	}
	if len(declared) != 0 {
		var extra []string
		for module := range declared {
			extra = append(extra, module)
		}
		sort.Strings(extra)
		t.Errorf("stdlib module policy names missing from shipped source: %s", strings.Join(extra, ", "))
	}
}

func TestStdlibNamedRealizationsTellTruth(t *testing.T) {
	for _, policy := range StdlibModulePolicies() {
		if strings.HasSuffix(policy.Module, "/native") && policy.Realization != RealizationNative {
			t.Errorf("%s is named /native but declared %s", policy.Module, policy.Realization)
		}
		if strings.HasSuffix(policy.Module, "/ffi") && policy.Realization != RealizationFFI {
			t.Errorf("%s is named /ffi but declared %s", policy.Module, policy.Realization)
		}
	}
}

func TestPortableNativeRealizationsOwnBareDefaults(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	if err := registry.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatal(err)
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationNative || !strings.HasSuffix(policy.Module, "/native") {
			continue
		}
		bare := strings.TrimSuffix(policy.Module, "/native")
		_, canonical, ok := registry.Resolve(bare)
		if !ok {
			t.Errorf("portable native module %s has no bare default %s", policy.Module, bare)
			continue
		}
		if canonical != policy.Module {
			t.Errorf("bare default %s resolves to %s, want native %s", bare, canonical, policy.Module)
		}
	}
}

func TestPortableFFIHasNativeSemanticAuthority(t *testing.T) {
	policies := make(map[string]StdlibModulePolicy)
	for _, policy := range StdlibModulePolicies() {
		policies[policy.Module] = policy
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationFFI {
			continue
		}
		if policy.SemanticAuthority == "" {
			t.Errorf("portable FFI module %s has no native semantic authority", policy.Module)
			continue
		}
		authority, ok := policies[policy.SemanticAuthority]
		if !ok {
			t.Errorf("portable FFI module %s names missing authority %s", policy.Module, policy.SemanticAuthority)
			continue
		}
		if authority.Role != RolePortable || authority.Realization != RealizationNative {
			t.Errorf("portable FFI module %s authority %s is %s/%s, want portable/native",
				policy.Module, authority.Module, authority.Role, authority.Realization)
		}
	}
}

func TestPortableFFIDoesNotFallBackToNativeAuthority(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	if err := registry.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatal(err)
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationFFI {
			continue
		}
		path, canonical, ok := registry.Resolve(policy.Module)
		if !ok || canonical != policy.Module {
			t.Errorf("FFI module %s does not resolve canonically", policy.Module)
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Error(err)
			continue
		}
		source := string(data)
		authorityRE := regexp.MustCompile(`(?m)^[[:space:]]*(?:let[[:space:]]+[_A-Za-z][_A-Za-z0-9]*[[:space:]]*=[[:space:]]*)?(?:import|use)[[:space:]]*\([[:space:]]*"` + regexp.QuoteMeta(policy.SemanticAuthority) + `"`)
		if authorityRE.MatchString(source) {
			t.Errorf("FFI module %s falls back to native semantic authority %s", policy.Module, policy.SemanticAuthority)
		}
	}
}

func TestPortableNativeSourcesDoNotUseProviderImplementations(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	if err := registry.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatal(err)
	}
	providerCallRE := regexp.MustCompile(`\b(_[A-Za-z][A-Za-z0-9_]*)\s*\(`)
	ffiImportRE := regexp.MustCompile(`\b(?:import|use)\s*\(\s*"[^"]+/ffi"`)
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationNative {
			continue
		}
		path, canonical, ok := registry.Resolve(policy.Module)
		if !ok || canonical != policy.Module {
			t.Errorf("native module %s does not resolve canonically", policy.Module)
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Error(err)
			continue
		}
		source := string(data)
		if ffiImportRE.MatchString(source) {
			t.Errorf("native module %s imports an FFI implementation", policy.Module)
		}
		for _, match := range providerCallRE.FindAllStringSubmatch(source, -1) {
			name := match[1]
			if role, ok := primitives.RoleOf(name); ok && role == primitives.RoleProvider {
				t.Errorf("native module %s calls provider primitive %s", policy.Module, name)
			}
		}
	}
}

func TestPortableNativeSourcesHaveNoTransitiveFFI(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	g := testRegistryGrammar(t)
	if err := registry.Scan(g); err != nil {
		t.Fatal(err)
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationNative {
			continue
		}
		path, canonical, ok := registry.Resolve(policy.Module)
		if !ok || canonical != policy.Module {
			t.Errorf("native module %s does not resolve canonically", policy.Module)
			continue
		}
		uses, err := FFIUsage(g, path, []string{root})
		if err != nil {
			t.Errorf("native module %s FFI graph: %v", policy.Module, err)
			continue
		}
		if len(uses) != 0 {
			t.Errorf("native module %s reaches FFI implementations transitively: %v", policy.Module, uses)
		}
	}
}

func TestPreludeDoesNotUseProviderImplementations(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "prelude", "prelude.ai"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	providerCallRE := regexp.MustCompile(`\b(_[A-Za-z][A-Za-z0-9_]*)\s*\(`)
	for _, match := range providerCallRE.FindAllStringSubmatch(string(data), -1) {
		name := match[1]
		if role, ok := primitives.RoleOf(name); ok && role == primitives.RoleProvider {
			t.Errorf("prelude calls provider primitive %s; provider library behavior must remain explicit", name)
		}
	}
}

func TestPortableFFIExportsAreProviderBacked(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	g := testRegistryGrammar(t)
	if err := registry.Scan(g); err != nil {
		t.Fatal(err)
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationFFI {
			continue
		}
		path, canonical, ok := registry.Resolve(policy.Module)
		if !ok || canonical != policy.Module {
			t.Errorf("FFI module %s does not resolve canonically", policy.Module)
			continue
		}
		info := analyzePolicySource(t, g, path)
		providerBacked := make(map[string]bool)
		for name, fn := range info.Functions {
			for _, call := range fn.Calls {
				if role, ok := primitives.RoleOf(call); ok && role == primitives.RoleProvider {
					providerBacked[name] = true
					break
				}
			}
		}
		changed := true
		for changed {
			changed = false
			for name, fn := range info.Functions {
				if providerBacked[name] {
					continue
				}
				for _, call := range fn.Calls {
					if providerBacked[call] {
						providerBacked[name] = true
						changed = true
						break
					}
				}
			}
		}
		for _, name := range info.Exports {
			if _, isFunction := info.Functions[name]; isFunction && !providerBacked[name] {
				t.Errorf("FFI module %s export %s has no provider-backed implementation path", policy.Module, name)
			}
		}
	}
}

func TestAccelerationExportsAndSignaturesMatchNative(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	g := testRegistryGrammar(t)
	if err := registry.Scan(g); err != nil {
		t.Fatal(err)
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationFFI {
			continue
		}
		ffiPath, _, ok := registry.Resolve(policy.Module)
		if !ok {
			t.Errorf("missing FFI module %s", policy.Module)
			continue
		}
		nativePath, _, ok := registry.Resolve(policy.SemanticAuthority)
		if !ok {
			t.Errorf("missing native authority %s", policy.SemanticAuthority)
			continue
		}
		ffiInfo := analyzePolicySource(t, g, ffiPath)
		nativeInfo := analyzePolicySource(t, g, nativePath)
		if got, want := sortedStrings(ffiInfo.Exports), sortedStrings(nativeInfo.Exports); strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Errorf("%s exports %v; native %s exports %v", policy.Module, got, policy.SemanticAuthority, want)
			continue
		}
		for _, name := range nativeInfo.Exports {
			nativeFn, nativeOK := nativeInfo.Functions[name]
			ffiFn, ffiOK := ffiInfo.Functions[name]
			if !nativeOK || !ffiOK {
				continue
			}
			if strings.Join(nativeFn.Parameters, "\x00") != strings.Join(ffiFn.Parameters, "\x00") || nativeFn.Rest != ffiFn.Rest {
				t.Errorf("%s.%s signature differs from %s.%s", policy.Module, name, policy.SemanticAuthority, name)
			}
		}
	}
}

func TestAccelerationHelpAndDocSurfaceMatchesNative(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "..", "lib"))
	registry := NewModuleRegistry([]string{root})
	if err := registry.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatal(err)
	}
	for _, policy := range StdlibModulePolicies() {
		if policy.Role != RolePortable || policy.Realization != RealizationFFI {
			continue
		}
		ffiPath, _, ok := registry.Resolve(policy.Module)
		if !ok {
			t.Errorf("missing FFI module %s", policy.Module)
			continue
		}
		nativePath, _, ok := registry.Resolve(policy.SemanticAuthority)
		if !ok {
			t.Errorf("missing native authority %s", policy.SemanticAuthority)
			continue
		}
		ffiHelp := parsePolicyHelp(t, ffiPath)
		nativeHelp := parsePolicyHelp(t, nativePath)
		if got, want := sortedMapKeys(ffiHelp), sortedMapKeys(nativeHelp); strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Errorf("%s help entries %v; native %s entries %v", policy.Module, got, policy.SemanticAuthority, want)
			continue
		}
		for name, nativeEntry := range nativeHelp {
			if ffiEntry := ffiHelp[name]; ffiEntry.Template != nativeEntry.Template {
				t.Errorf("%s.%s help template %q; native %s.%s is %q",
					policy.Module, name, ffiEntry.Template, policy.SemanticAuthority, name, nativeEntry.Template)
			}
		}
		ffiDocs := parsePolicyDocs(t, ffiPath)
		nativeDocs := parsePolicyDocs(t, nativePath)
		if got, want := sortedMapKeys(ffiDocs), sortedMapKeys(nativeDocs); strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Errorf("%s doc entries %v; native %s entries %v", policy.Module, got, policy.SemanticAuthority, want)
		}
	}
}

func parsePolicyHelp(t *testing.T, aiPath string) map[string]help.FuncEntry {
	t.Helper()
	path := strings.TrimSuffix(aiPath, ".ai") + ".help"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := help.ParseHelpFile(path, string(data))
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func parsePolicyDocs(t *testing.T, aiPath string) map[string]help.DocEntry {
	t.Helper()
	path := strings.TrimSuffix(aiPath, ".ai") + ".doc"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := help.ParseDocFile(path, string(data))
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func sortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func analyzePolicySource(t *testing.T, g *grammar.Grammar, path string) SourceInfo {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := AnalyzeSource(g, path, string(data))
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
