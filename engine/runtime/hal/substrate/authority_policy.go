package substrate

import (
	"path/filepath"
	"strings"

	"aiki/engine/semantics/value"
)

// authorityPolicy records the raw runtime bindings actually referenced by each
// trusted Aiki source. Paths identify canonical project modules, but path
// topology does not itself confer privilege. AuthorityForSource translates any
// host binding in this table to its canonical HAL authority identity.
var authorityPolicy = map[string][]string{
	"engine/runtime/prelude/prelude.ai": {
		"_append", "_apply", "_channel", "_chr", "_delete", "_doc", "_empty", "_equal",
		"_first", "_help", "_inspect", "_length", "_load", "_make_shaped_list", "_ord",
		"_prepend", "_print", "_quit", "_read", "_recv", "_reset", "_rest", "_send", "_shape",
		"_spawn", "_stack_limit", "_to_decimal", "_to_number", "_to_str", "_to_symbol", "_truncate", "_type",
	},
	"lib/bits/bits.ai":       {"_bits_and", "_bits_not", "_bits_or", "_bits_shl", "_bits_shr", "_bits_xor"},
	"lib/bytes/ffi.ai":       {"_bytes_get", "_bytes_length", "_bytes_new", "_bytes_slice", "_bytes_to_str", "_str_to_bytes"},
	"lib/bytes/native.ai":    {"_bytes_to_str_pure", "_make_shaped_list"},
	"lib/canvas/canvas.ai":   {"_canvas", "_canvas_alive", "_canvas_command", "_canvas_height", "_canvas_width", "_destroy"},
	"lib/file/file.ai":       {"_file_close", "_file_copy", "_file_delete", "_file_exists", "_file_list", "_file_mkdir", "_file_mkdir_all", "_file_open", "_file_read_at", "_file_read_bytes", "_file_read_line", "_file_read_text", "_file_remove_all", "_file_rename", "_file_size", "_file_stat", "_file_temp", "_file_temp_dir", "_file_write_at", "_file_write_bytes", "_file_write_text"},
	"lib/hash/ffi.ai":        {"_modulo"},
	"lib/path/path.ai":       {"_path_separator"},
	"lib/math/ffi.ai":        {"_ceil", "_cos_inexact", "_floor", "_modulo", "_sin_inexact", "_sqrt_inexact"},
	"lib/profile/profile.ai": {"_profile_counts", "_profile_experiment", "_profile_measure"},
	"lib/random/random.ai":   {"_random", "_seed"},
	"lib/regex/ffi.ai":       {"_regex_find", "_regex_find_all", "_regex_match", "_regex_replace", "_regex_replace_first", "_regex_split"},
	"lib/selfhost/bootstrap.ai": {
		"_floor", "_ceil", "_truncate", "_modulo", "_sqrt_inexact", "_cos_inexact", "_sin_inexact", "_seed", "_random",
		"_upper", "_lower", "_chars", "_upper_rune", "_lower_rune", "_bytes_new", "_bytes_length", "_bytes_get", "_bytes_slice",
		"_str_to_bytes", "_bytes_to_str", "_bytes_to_str_pure", "_regex_match", "_regex_find", "_regex_find_all", "_regex_replace",
		"_regex_replace_first", "_regex_split", "_file_open", "_file_read_text", "_file_read_bytes", "_file_read_line", "_file_write_text",
		"_file_write_bytes", "_file_close", "_file_exists", "_file_delete", "_file_list", "_file_read_at", "_file_write_at", "_file_stat", "_file_rename", "_file_mkdir", "_file_mkdir_all", "_file_remove_all", "_file_temp", "_file_temp_dir", "_file_copy", "_file_size", "_sleep", "_after", "_time_now",
		"_system_args", "_system_env", "_path_separator", "_system_cwd", "_system_chdir", "_system_exec", "_module_roots", "_store_new", "_store_get", "_store_set", "_store_length", "_bits_and", "_bits_or",
		"_bits_xor", "_bits_not", "_bits_shl", "_bits_shr", "_apply", "_shape", "_make_shaped_list", "_canvas", "_canvas_command", "_destroy", "_canvas_width",
		"_canvas_height", "_canvas_alive", "_profile_counts", "_profile_measure", "_profile_experiment", "_test_equal",
		"_test_not_equal", "_test_true", "_test_false", "_test_error", "_test_not_error", "_test_run", "_test_faults",
	},
	"lib/store/store.ai":   {"_store_get", "_store_length", "_store_new", "_store_set"},
	"lib/string/string.ai": {"_chars", "_lower", "_lower_rune", "_upper", "_upper_rune"},
	"lib/system/system.ai": {"_system_args", "_system_chdir", "_system_cwd", "_system_env", "_system_exec"},
	"lib/test/test.ai":     {"_test_equal", "_test_error", "_test_false", "_test_faults", "_test_not_equal", "_test_not_error", "_test_run", "_test_true"},
	"lib/time/time.ai":     {"_after", "_sleep", "_time_now"},
	"lib/turtle/simple.ai": {"_canvas", "_canvas_alive", "_canvas_command", "_canvas_width", "_cos_inexact", "_destroy", "_sin_inexact", "_truncate"},
	"lib/turtle/turtle.ai": {"_canvas_command", "_canvas_height", "_canvas_width", "_cos_inexact", "_sin_inexact"},
}

// AuthorityForSource returns only the grants explicitly declared for a
// canonical trusted source. Host bindings are translated to canonical HAL
// identities; non-host primitives retain their implementation names. Merely
// residing beneath lib/ confers nothing.
func (g *GoRuntime) AuthorityForSource(path string) value.Authority {
	normalized := filepath.ToSlash(filepath.Clean(path))
	for suffix, bindings := range authorityPolicy {
		if normalized == suffix || strings.HasSuffix(normalized, "/"+suffix) {
			grants := make([]string, 0, len(bindings))
			for _, name := range bindings {
				grants = append(grants, g.authorityKey(name))
			}
			return value.NewAuthority(grants...)
		}
	}
	return value.NoAuthority()
}
