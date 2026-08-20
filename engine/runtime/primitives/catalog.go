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
	RoleRuntime   Role = "runtime"
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

	// Constitutive language/value/runtime primitives.
	"_first":                       RoleRuntime,
	"_rest":                        RoleRuntime,
	"_length":                      RoleRuntime,
	"_prepend":                     RoleRuntime,
	"_append":                      RoleRuntime,
	"_empty":                       RoleRuntime,
	"_range":                       RoleRuntime,
	"_type":                        RoleRuntime,
	"_stack_limit":                 RoleRuntime,
	"_inspect":                     RoleRuntime,
	"_equal":                       RoleRuntime,
	"_ord":                         RoleRuntime,
	"_chr":                         RoleRuntime,
	"_floor":                       RoleProvider,
	"_ceil":                        RoleProvider,
	"_truncate":                    RoleRuntime,
	"_modulo":                      RoleProvider,
	"_bytes_new":                   RoleProvider,
	"_bytes_length":                RoleProvider,
	"_bytes_get":                   RoleProvider,
	"_bytes_slice":                 RoleProvider,
	"_str_to_bytes":                RoleProvider,
	"_bytes_to_str":                RoleProvider,
	"_bytes_to_str_pure":           RoleProvider,
	"_bytes_digits_from_text":      RoleProvider,
	"_bytes_digits_to_text":        RoleProvider,
	"_shape":                       RoleRuntime,
	"_make_shaped_list":            RoleRuntime,
	"_to_str":                      RoleRuntime,
	"_to_decimal":                  RoleRuntime,
	"_to_number":                   RoleRuntime,
	"_to_symbol":                   RoleRuntime,
	"_store_new":                   RoleRuntime,
	"_store_get":                   RoleRuntime,
	"_store_set":                   RoleRuntime,
	"_store_length":                RoleRuntime,
	"_store_snapshot":              RoleRuntime,
	"_fixed_store_new_byte":        RoleProvider,
	"_fixed_store_new_word":        RoleProvider,
	"_fixed_store_new_addr18":      RoleProvider,
	"_fixed_store_new_counter":     RoleProvider,
	"_fixed_store_get":             RoleProvider,
	"_fixed_store_set":             RoleProvider,
	"_fixed_store_length":          RoleProvider,
	"_fixed_store_snapshot":        RoleProvider,
	"_fixed_store_word_read_addr":  RoleProvider,
	"_fixed_store_word_write_addr": RoleProvider,
	"_fixed_store_byte_read_addr":  RoleProvider,
	"_fixed_store_byte_write_addr": RoleProvider,
	"_fixed_store_counter_get":     RoleProvider,
	"_fixed_store_counter_set":     RoleProvider,
	"_fixed_store_counter_add":     RoleProvider,
	"_machine_word":                RoleProvider,
	"_machine_byte":                RoleProvider,
	"_machine_addr18":              RoleProvider,
	"_machine_number":              RoleProvider,
	"_machine_same":                RoleProvider,
	"_machine_word_add":            RoleProvider,
	"_machine_word_sub":            RoleProvider,
	"_machine_word_and":            RoleProvider,
	"_machine_word_or":             RoleProvider,
	"_machine_word_xor":            RoleProvider,
	"_machine_word_not":            RoleProvider,
	"_machine_word_shl":            RoleProvider,
	"_machine_word_shr":            RoleProvider,
	"_machine_word_extract":        RoleProvider,
	"_machine_word_lt":             RoleProvider,
	"_machine_word_gt":             RoleProvider,
	"_machine_word_zero":           RoleProvider,
	"_machine_word_sign":           RoleProvider,
	"_machine_word_low_byte":       RoleProvider,
	"_machine_word_high_byte":      RoleProvider,
	"_machine_word_with_low_byte":  RoleProvider,
	"_machine_word_with_high_byte": RoleProvider,
	"_machine_byte_to_word":        RoleProvider,
	"_machine_byte_sign_word":      RoleProvider,
	"_machine_byte_zero":           RoleProvider,
	"_machine_addr_add":            RoleProvider,
	"_machine_addr_even":           RoleProvider,
	"_machine_word_field":          RoleProvider,
	"_machine_word_mask":           RoleProvider,
	"_machine_word_eq_number":      RoleProvider,
	"_machine_word_lt_number":      RoleProvider,
	"_machine_word_ge_number":      RoleProvider,
	"_machine_word_add_small":      RoleProvider,
	"_machine_word_sub_small":      RoleProvider,
	"_machine_word_add_carry":      RoleProvider,
	"_machine_word_sub_borrow":     RoleProvider,
	"_machine_word_bit":            RoleProvider,
	"_machine_word_any_mask":       RoleProvider,
	"_machine_word_set_bit":        RoleProvider,
	"_machine_zero":                RoleProvider,
	"_machine_negative":            RoleProvider,
	"_machine_addr_from_word":      RoleProvider,
	"_machine_addr_lt_number":      RoleProvider,
	"_machine_addr_ge_number":      RoleProvider,
	"_machine_add":                 RoleProvider,
	"_machine_sub":                 RoleProvider,
	"_machine_and":                 RoleProvider,
	"_machine_or":                  RoleProvider,
	"_machine_xor":                 RoleProvider,
	"_machine_not":                 RoleProvider,
	"_machine_add_small":           RoleProvider,
	"_machine_sub_small":           RoleProvider,
	"_machine_lt":                  RoleProvider,
	"_machine_gt":                  RoleProvider,
	"_machine_add_carry_generic":   RoleProvider,
	"_machine_sub_borrow_generic":  RoleProvider,
	"_machine_shr1":                RoleProvider,
	"_machine_shl1":                RoleProvider,
	"_machine_low_bit":             RoleProvider,
	"_machine_set_high_bit":        RoleProvider,
	"_bits_and":                    RoleProvider,
	"_bits_or":                     RoleProvider,
	"_bits_xor":                    RoleProvider,
	"_bits_not":                    RoleProvider,
	"_bits_shl":                    RoleProvider,
	"_bits_shr":                    RoleProvider,
	"_hash_code":                   RoleProvider,
	"_hash_new":                    RoleProvider,
	"_hash_get":                    RoleProvider,
	"_hash_put":                    RoleProvider,
	"_hash_has":                    RoleProvider,
	"_hash_del":                    RoleProvider,
	"_hash_keys":                   RoleProvider,
	"_hash_values":                 RoleProvider,

	// Native/FFI library providers.
	"_sqrt_inexact":         RoleProvider,
	"_cos_inexact":          RoleProvider,
	"_sin_inexact":          RoleProvider,
	"_upper":                RoleProvider,
	"_lower":                RoleProvider,
	"_chars":                RoleProvider,
	"_upper_rune":           RoleProvider,
	"_lower_rune":           RoleProvider,
	"_string_substring":     RoleProvider,
	"_string_split":         RoleProvider,
	"_string_index_of":      RoleProvider,
	"_string_last_index_of": RoleProvider,
	"_string_contains":      RoleProvider,
	"_string_starts_with":   RoleProvider,
	"_string_ends_with":     RoleProvider,
	"_string_join":          RoleProvider,
	"_string_replace":       RoleProvider,
	"_string_replace_first": RoleProvider,
	"_string_repeat":        RoleProvider,
	"_string_reverse":       RoleProvider,
	"_string_trim":          RoleProvider,
	"_string_trim_start":    RoleProvider,
	"_string_trim_end":      RoleProvider,
	"_string_compare":       RoleProvider,
	"_string_is_whitespace": RoleProvider,
	"_string_is_digit":      RoleProvider,
	"_string_is_alpha":      RoleProvider,
	"_string_is_upper":      RoleProvider,
	"_string_is_lower":      RoleProvider,
	"_string_is_alnum":      RoleProvider,
	"_string_is_numeric":    RoleProvider,
	"_string_is_alphabetic": RoleProvider,
	"_regex_match":          RoleProvider,
	"_regex_find":           RoleProvider,
	"_regex_find_all":       RoleProvider,
	"_regex_replace":        RoleProvider,
	"_regex_replace_first":  RoleProvider,
	"_regex_split":          RoleProvider,

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
