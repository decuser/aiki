// Package primitives owns the architectural classification of Aiki runtime
// primitives. Concrete substrates bind implementations to these stable names;
// the role classification is not substrate metadata.
package primitives

import (
	"sort"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// Role identifies why a runtime primitive exists.
type Role string

const (
	RoleIntrinsic Role = "intrinsic"
	RoleNative    Role = "native"
	RoleProvider  Role = "provider"
	RoleHost      Role = "host"
	RoleService   Role = "service"
)

var nonHALRoles = map[string]Role{
	// Evaluator/language intrinsics.
	"_apply":   RoleIntrinsic,
	"_import":  RoleIntrinsic,
	"_use":     RoleIntrinsic,
	"_export":  RoleIntrinsic,
	"_load":    RoleIntrinsic,
	"_spawn":   RoleIntrinsic,
	"_channel": RoleIntrinsic,
	"_send":    RoleIntrinsic,
	"_recv":    RoleIntrinsic,

	// Language/value primitives implemented natively.
	"_first":             RoleNative,
	"_rest":              RoleNative,
	"_length":            RoleNative,
	"_prepend":           RoleNative,
	"_append":            RoleNative,
	"_empty":             RoleNative,
	"_range":             RoleNative,
	"_type":              RoleNative,
	"_stack_limit":       RoleNative,
	"_inspect":           RoleNative,
	"_equal":             RoleNative,
	"_ord":               RoleNative,
	"_chr":               RoleNative,
	"_floor":             RoleNative,
	"_ceil":              RoleNative,
	"_truncate":          RoleNative,
	"_modulo":            RoleNative,
	"_bytes_new":         RoleNative,
	"_bytes_length":      RoleNative,
	"_bytes_get":         RoleNative,
	"_bytes_slice":       RoleNative,
	"_str_to_bytes":      RoleNative,
	"_bytes_to_str":      RoleNative,
	"_bytes_to_str_pure": RoleNative,
	"_shape":             RoleNative,
	"_make_shaped_list":  RoleNative,
	"_to_str":            RoleNative,
	"_to_decimal":        RoleNative,
	"_to_number":         RoleNative,
	"_to_symbol":         RoleNative,
	"_store_new":         RoleNative,
	"_store_get":         RoleNative,
	"_store_set":         RoleNative,
	"_store_length":      RoleNative,
	"_bits_and":          RoleNative,
	"_bits_or":           RoleNative,
	"_bits_xor":          RoleNative,
	"_bits_not":          RoleNative,
	"_bits_shl":          RoleNative,
	"_bits_shr":          RoleNative,

	// Native/FFI library providers.
	"_sqrt_inexact":        RoleProvider,
	"_cos_inexact":         RoleProvider,
	"_sin_inexact":         RoleProvider,
	"_upper":               RoleProvider,
	"_lower":               RoleProvider,
	"_chars":               RoleProvider,
	"_upper_rune":          RoleProvider,
	"_lower_rune":          RoleProvider,
	"_regex_match":         RoleProvider,
	"_regex_find":          RoleProvider,
	"_regex_find_all":      RoleProvider,
	"_regex_replace":       RoleProvider,
	"_regex_replace_first": RoleProvider,
	"_regex_split":         RoleProvider,

	// Runtime/tooling/session services.
	"_module_roots":       RoleService,
	"_system_exit":        RoleService,
	"_system_has":         RoleService,
	"_system_require":     RoleService,
	"_profile_counts":     RoleService,
	"_profile_measure":    RoleService,
	"_profile_experiment": RoleService,
	"_quit":               RoleService,
	"_reset":              RoleService,
	"_delete":             RoleService,
	"_help":               RoleService,
	"_doc":                RoleService,
	"_test_equal":         RoleService,
	"_test_not_equal":     RoleService,
	"_test_true":          RoleService,
	"_test_false":         RoleService,
	"_test_error":         RoleService,
	"_test_not_error":     RoleService,
	"_test_run":           RoleService,
	"_test_faults":        RoleService,
}

// RoleOf returns the architectural role for a primitive. Canonical HAL host
// operations derive their host role from the HAL operation authority rather
// than being repeated in this catalog.
func RoleOf(name string) (Role, bool) {
	if _, ok := hal.OperationDefinitions()[name]; ok {
		return RoleHost, true
	}
	role, ok := nonHALRoles[name]
	return role, ok
}

// Definitions returns a copy of every runtime primitive role, including
// canonical HAL host operations.
func Definitions() map[string]Role {
	out := make(map[string]Role, len(nonHALRoles)+len(hal.OperationDefinitions()))
	for name, role := range nonHALRoles {
		out[name] = role
	}
	for name := range hal.OperationDefinitions() {
		out[name] = RoleHost
	}
	return out
}

// NamesForScope returns the architectural runtime names visible to tooling for
// a lexical scope. User code sees only the language intrinsics directly;
// trusted/prelude scopes also see the primitive namespace.
func NamesForScope(scope value.Scope) []string {
	set := map[string]bool{
		"import": true,
		"use":    true,
		"export": true,
	}
	if scope != value.ScopeUser {
		for name := range Definitions() {
			set[name] = true
		}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
