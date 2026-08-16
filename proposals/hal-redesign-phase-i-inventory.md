# HAL Redesign — Phase I Boundary Inventory

Baseline: `v0.4.0-alpha-27`

This inventory is source-derived from `register.go`, the substrate builtin implementations, `prelude.ai`, and production `lib/*/*.ai`. It records all 117 currently registered `_` primitives.

## Immediate finding

The current registry is broader than a host HAL. It contains true host effects, evaluator intrinsics, ordinary value operations, accelerators/FFI implementations, REPL/test tooling, observation machinery, and stateful resources in one namespace. The redesign therefore begins by separating *kind of boundary* before designing a replacement interface.

## Full inventory

| # | Current primitive | Candidate HAL identity | Aiki-side use | Substrate realization/provenance | Current classification | Source |
|---:|---|---|---|---|---|---|
| 1 | `_print` | `HAL.io.print` | print (engine/runtime/prelude/prelude.ai:37); println (engine/runtime/prelude/prelude.ai:41); println (engine/runtime/prelude/prelude.ai:42) | `go:io.Writer/fmt.Fprint` | host capability | `substrate/builtins_io.go` |
| 2 | `_read` | `HAL.io.read` | read (engine/runtime/prelude/prelude.ai:46); input (engine/runtime/prelude/prelude.ai:50) | `go:io.Reader/bufio.Reader` | host capability | `substrate/builtins_io.go` |
| 3 | `_first` | `HAL.list.first` | first (engine/runtime/prelude/prelude.ai:55) | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 4 | `_rest` | `HAL.list.rest` | rest (engine/runtime/prelude/prelude.ai:59) | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 5 | `_length` | `HAL.list.length` | length (engine/runtime/prelude/prelude.ai:63) | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 6 | `_prepend` | `HAL.list.prepend` | spawn (engine/runtime/prelude/prelude.ai:20); prepend (engine/runtime/prelude/prelude.ai:67) | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 7 | `_append` | `HAL.list.append` | append (engine/runtime/prelude/prelude.ai:71) | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 8 | `_empty` | `HAL.list.empty` | empty (engine/runtime/prelude/prelude.ai:75) | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 9 | `_range` | `HAL.list.range` | no production Aiki reference found | `go:value.List` | language/value primitive | `substrate/builtins_list.go` |
| 10 | `_type` | `HAL.value.type` | type (engine/runtime/prelude/prelude.ai:80) | `go:value model` | language/value primitive | `substrate/builtins_type.go` |
| 11 | `_stack_limit` | `HAL.value.stack_limit` | stack_limit (engine/runtime/prelude/prelude.ai:15) | `go:value model` | language/value primitive | `substrate/builtins_type.go` |
| 12 | `_inspect` | `HAL.value.inspect` | inspect (engine/runtime/prelude/prelude.ai:84) | `go:value model` | language/value primitive | `substrate/builtins_type.go` |
| 13 | `_equal` | `HAL.value.equal` | equal (engine/runtime/prelude/prelude.ai:88); is_error (engine/runtime/prelude/prelude.ai:104) | `go:value model` | language/value primitive | `substrate/builtins_type.go` |
| 14 | `_ord` | `HAL.value.ord` | ord (engine/runtime/prelude/prelude.ai:92) | `go:value model` | language/value primitive | `substrate/builtins_type.go` |
| 15 | `_chr` | `HAL.value.chr` | chr (engine/runtime/prelude/prelude.ai:96) | `go:value model` | language/value primitive | `substrate/builtins_convert.go` |
| 16 | `_floor` | `HAL.math.floor` | math.floor (lib/math/ffi.ai:7) | `go:math/big exact arithmetic` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 17 | `_ceil` | `HAL.math.ceil` | math.ceil (lib/math/ffi.ai:11) | `go:math/big exact arithmetic` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 18 | `_truncate` | `HAL.math.truncate` | truncate (engine/runtime/prelude/prelude.ai:149); turtle.position (lib/turtle/simple.ai:191) | `go:math/big exact arithmetic` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 19 | `_modulo` | `HAL.math.modulo` | math.modulo (lib/math/ffi.ai:15); hash._bucket_index (lib/hash/ffi.ai:57) | `go:math/big exact arithmetic` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 20 | `_sqrt_inexact` | `HAL.math.sqrt_inexact` | math.sqrt (lib/math/ffi.ai:27) | `go:math float64 bridge` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 21 | `_cos_inexact` | `HAL.math.cos_inexact` | turtle.forward (lib/turtle/simple.ai:107); turtle.forward (lib/turtle/turtle.ai:79); math.cos (lib/math/ffi.ai:23) | `go:math float64 bridge` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_trig.go` |
| 22 | `_sin_inexact` | `HAL.math.sin_inexact` | turtle.forward (lib/turtle/simple.ai:106); turtle.forward (lib/turtle/turtle.ai:78); math.sin (lib/math/ffi.ai:19) | `go:math float64 bridge` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_trig.go` |
| 23 | `_seed` | `HAL.math.seed` | random.seed (lib/random/random.ai:5) | `go:math/rand.New` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 24 | `_random` | `HAL.math.random` | random.random (lib/random/random.ai:9) | `go:math/rand.Int63n` | mixed: value primitive/accelerator/nondeterminism | `substrate/builtins_math.go` |
| 25 | `_bytes_new` | `HAL.bytes.new` | bytes.bytes_new (lib/bytes/ffi.ai:18) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 26 | `_bytes_length` | `HAL.bytes.length` | bytes.bytes_length (lib/bytes/ffi.ai:28) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 27 | `_bytes_get` | `HAL.bytes.get` | bytes.bytes_get (lib/bytes/ffi.ai:23) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 28 | `_bytes_slice` | `HAL.bytes.slice` | bytes.bytes_slice (lib/bytes/ffi.ai:33) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 29 | `_str_to_bytes` | `HAL.bytes.str_to_bytes` | bytes.str_to_bytes (lib/bytes/ffi.ai:8) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 30 | `_bytes_to_str` | `HAL.bytes.to_str` | bytes.bytes_to_str (lib/bytes/ffi.ai:13) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 31 | `_bytes_to_str_pure` | `HAL.bytes.to_str_pure` | bytes.bytes_to_str (lib/bytes/native.ai:52) | `go:value.Bytes/string conversion` | language/value primitive | `substrate/builtins_bytes.go` |
| 32 | `_sleep` | `HAL.time.sleep` | time.sleep (lib/time/time.ai:5) | `go:time.Sleep` | host capability | `substrate/builtins_trig.go` |
| 33 | `_after` | `HAL.time.after` | time.after (lib/time/time.ai:9) | `go:time.AfterFunc` | host capability | `substrate/builtins_time.go` |
| 34 | `_system_args` | `HAL.system.args` | system.args (lib/system/system.ai:5) | `go:runtime-owned programArgs` | mixed: host/runtime | `substrate/builtins_system.go` |
| 35 | `_system_env` | `HAL.system.env` | system.env (lib/system/system.ai:9) | `go:os.LookupEnv` | mixed: host/runtime | `substrate/builtins_system.go` |
| 36 | `_module_roots` | `HAL.system.module_roots` | no production Aiki reference found | `go:ModuleRegistry.Roots` | mixed: host/runtime | `substrate/builtins_system.go` |
| 37 | `_canvas` | `HAL.canvas.canvas` | canvas.canvas (lib/canvas/canvas.ai:5); turtle.new (lib/turtle/simple.ai:84) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 38 | `_dot` | `HAL.canvas.dot` | canvas.dot (lib/canvas/canvas.ai:17) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 39 | `_line` | `HAL.canvas.line` | canvas.line (lib/canvas/canvas.ai:21); turtle.forward (lib/turtle/simple.ai:113); turtle.forward (lib/turtle/turtle.ai:85); turtle.home (lib/turtle/turtle.ai:118) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 40 | `_rect` | `HAL.canvas.rect` | canvas.rect (lib/canvas/canvas.ai:25) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 41 | `_fill_rect` | `HAL.canvas.fill_rect` | canvas.fill_rect (lib/canvas/canvas.ai:29) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 42 | `_circle` | `HAL.canvas.circle` | canvas.circle (lib/canvas/canvas.ai:33) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 43 | `_fill_circle` | `HAL.canvas.fill_circle` | canvas.fill_circle (lib/canvas/canvas.ai:37) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 44 | `_arc` | `HAL.canvas.arc` | canvas.arc (lib/canvas/canvas.ai:69) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 45 | `_clear` | `HAL.canvas.clear` | canvas.clear (lib/canvas/canvas.ai:41); turtle.new (lib/turtle/simple.ai:87); turtle.clear (lib/turtle/simple.ai:185); turtle.clear (lib/turtle/turtle.ai:126) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 46 | `_destroy` | `HAL.canvas.destroy` | canvas.destroy (lib/canvas/canvas.ai:45); turtle.new (lib/turtle/simple.ai:81) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 47 | `_set_bg` | `HAL.canvas.set_bg` | canvas.set_bg (lib/canvas/canvas.ai:49); canvas.set_bg_rgb (lib/canvas/canvas.ai:57); turtle.background (lib/turtle/simple.ai:162) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 48 | `_set_fg` | `HAL.canvas.set_fg` | canvas.set_fg (lib/canvas/canvas.ai:53); canvas.set_fg_rgb (lib/canvas/canvas.ai:61) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 49 | `_pen_size` | `HAL.canvas.pen_size` | canvas.pen_size (lib/canvas/canvas.ai:65); turtle.pen_size (lib/turtle/simple.ai:149); turtle.pen_size (lib/turtle/simple.ai:151) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas.go` |
| 50 | `_canvas_width` | `HAL.canvas.width` | canvas.width (lib/canvas/canvas.ai:9); turtle.new (lib/turtle/simple.ai:73); turtle.turtle (lib/turtle/turtle.ai:14) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas_accessors.go` |
| 51 | `_canvas_height` | `HAL.canvas.height` | canvas.height (lib/canvas/canvas.ai:13); turtle.turtle (lib/turtle/turtle.ai:15) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas_accessors.go` |
| 52 | `_canvas_alive` | `HAL.canvas.alive` | turtle.new (lib/turtle/simple.ai:72); turtle.new (lib/turtle/simple.ai:80) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas_accessors.go` |
| 53 | `_set_turtle` | `HAL.canvas.set_turtle` | turtle._update_turtle (lib/turtle/simple.ai:54) | `go:canvas session/IPC/Ebiten` | host resource/protocol pressure | `substrate/builtins_canvas_accessors.go` |
| 54 | `_shape` | `HAL.value.shape` | shape (engine/runtime/prelude/prelude.ai:100); is_error (engine/runtime/prelude/prelude.ai:104) | `go:value model` | language/value primitive | `substrate/builtins_convert.go` |
| 55 | `_make_shaped_list` | `HAL.value.make_shaped_list` | shaped (engine/runtime/prelude/prelude.ai:124); bytes._make_bytes (lib/bytes/native.ai:11) | `go:value model` | language/value primitive | `substrate/builtins_list.go` |
| 56 | `_to_str` | `HAL.value.to_str` | to_str (engine/runtime/prelude/prelude.ai:108) | `go:value model` | language/value primitive | `substrate/builtins_convert.go` |
| 57 | `_to_decimal` | `HAL.value.to_decimal` | to_decimal (engine/runtime/prelude/prelude.ai:112) | `go:value model` | language/value primitive | `substrate/builtins_convert.go` |
| 58 | `_to_number` | `HAL.value.to_number` | to_number (engine/runtime/prelude/prelude.ai:116) | `go:value model` | language/value primitive | `substrate/builtins_convert.go` |
| 59 | `_to_symbol` | `HAL.value.to_symbol` | to_symbol (engine/runtime/prelude/prelude.ai:120) | `go:value model` | language/value primitive | `substrate/builtins_convert.go` |
| 60 | `_upper` | `HAL.string.upper` | string.upper (lib/string/string.ai:340) | `go:strings/unicode` | library accelerator | `substrate/builtins_convert.go` |
| 61 | `_lower` | `HAL.string.lower` | string.lower (lib/string/string.ai:345) | `go:strings/unicode` | library accelerator | `substrate/builtins_convert.go` |
| 62 | `_chars` | `HAL.string.chars` | string.chars (lib/string/string.ai:245) | `go:strings/unicode` | library accelerator | `substrate/builtins_convert.go` |
| 63 | `_upper_rune` | `HAL.string.upper_rune` | string.upper_rune (lib/string/string.ai:350) | `go:strings/unicode` | library accelerator | `substrate/builtins_convert.go` |
| 64 | `_lower_rune` | `HAL.string.lower_rune` | string.lower_rune (lib/string/string.ai:355) | `go:strings/unicode` | library accelerator | `substrate/builtins_convert.go` |
| 65 | `_regex_match` | `HAL.regex.match` | regex.matches (lib/regex/ffi.ai:15) | `go:regexp` | library accelerator | `substrate/builtins_regex.go` |
| 66 | `_regex_find` | `HAL.regex.find` | regex.find (lib/regex/ffi.ai:21) | `go:regexp` | library accelerator | `substrate/builtins_regex.go` |
| 67 | `_regex_find_all` | `HAL.regex.find_all` | regex.find_all (lib/regex/ffi.ai:27) | `go:regexp` | library accelerator | `substrate/builtins_regex.go` |
| 68 | `_regex_replace` | `HAL.regex.replace` | regex.replace (lib/regex/ffi.ai:34) | `go:regexp` | library accelerator | `substrate/builtins_regex.go` |
| 69 | `_regex_replace_first` | `HAL.regex.replace_first` | regex.replace_first (lib/regex/ffi.ai:40) | `go:regexp` | library accelerator | `substrate/builtins_regex.go` |
| 70 | `_regex_split` | `HAL.regex.split` | regex.split (lib/regex/ffi.ai:46) | `go:regexp` | library accelerator | `substrate/builtins_regex.go` |
| 71 | `_file_open` | `HAL.file.open` | file.open (lib/file/file.ai:5) | `go:os.Open/os.Create/os.OpenFile` | host capability | `substrate/builtins_file.go` |
| 72 | `_file_read_text` | `HAL.file.read_text` | file.read_text (lib/file/file.ai:9) | `go:io.ReadAll(*os.File)` | host capability | `substrate/builtins_file.go` |
| 73 | `_file_read_bytes` | `HAL.file.read_bytes` | file.read_bytes (lib/file/file.ai:13) | `go:io.ReadAll(*os.File)` | host capability | `substrate/builtins_file.go` |
| 74 | `_file_read_line` | `HAL.file.read_line` | file.read_line (lib/file/file.ai:17) | `go:bufio.Reader.ReadString` | host capability | `substrate/builtins_file.go` |
| 75 | `_file_write_text` | `HAL.file.write_text` | file.write_text (lib/file/file.ai:21) | `go:os.File.WriteString` | host capability | `substrate/builtins_file.go` |
| 76 | `_file_write_bytes` | `HAL.file.write_bytes` | file.write_bytes (lib/file/file.ai:25) | `go:os.File.Write` | host capability | `substrate/builtins_file.go` |
| 77 | `_file_close` | `HAL.file.close` | file.close (lib/file/file.ai:29) | `go:os.File.Close` | host capability | `substrate/builtins_file.go` |
| 78 | `_file_exists` | `HAL.file.exists` | file.exists (lib/file/file.ai:33) | `go:os.Stat` | host capability | `substrate/builtins_file.go` |
| 79 | `_file_delete` | `HAL.file.delete` | file.delete (lib/file/file.ai:37) | `go:os.Remove` | host capability | `substrate/builtins_file.go` |
| 80 | `_file_list` | `HAL.file.list` | file.list (lib/file/file.ai:41) | `go:os.ReadDir` | host capability | `substrate/builtins_file.go` |
| 81 | `_file_read_at` | `HAL.file.read_at` | file.read_at (lib/file/file.ai:45) | `go:os.File.ReadAt` | host capability | `substrate/builtins_file.go` |
| 82 | `_file_write_at` | `HAL.file.write_at` | file.write_at (lib/file/file.ai:49) | `go:os.File.WriteAt` | host capability | `substrate/builtins_file.go` |
| 83 | `_apply` | `HAL.language.apply` | apply (engine/runtime/prelude/prelude.ai:7); spawn (engine/runtime/prelude/prelude.ai:20); canvas.dot (lib/canvas/canvas.ai:17); canvas.line (lib/canvas/canvas.ai:21); canvas.rect (lib/canvas/canvas.ai:25); canvas.fill_rect (lib/canvas/canvas.ai:29); canvas.circle (lib/canvas/canvas.ai:33); canvas.fill_circle (lib/canvas/canvas.ai:37); canvas.arc (lib/canvas/canvas.ai:69) | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_intrinsic.go` |
| 84 | `_import` | `HAL.language.import` | no production Aiki reference found | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_intrinsic.go` |
| 85 | `_use` | `HAL.language.use` | no production Aiki reference found | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_intrinsic.go` |
| 86 | `_export` | `HAL.language.export` | no production Aiki reference found | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_intrinsic.go` |
| 87 | `_load` | `HAL.language.load` | load (engine/runtime/prelude/prelude.ai:11) | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_intrinsic.go` |
| 88 | `_spawn` | `HAL.language.spawn` | spawn (engine/runtime/prelude/prelude.ai:20) | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_intrinsic.go` |
| 89 | `_channel` | `HAL.language.channel` | channel (engine/runtime/prelude/prelude.ai:24) | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_concurrency.go` |
| 90 | `_send` | `HAL.language.send` | send (engine/runtime/prelude/prelude.ai:28) | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_concurrency.go` |
| 91 | `_recv` | `HAL.language.recv` | recv (engine/runtime/prelude/prelude.ai:32) | `go:evaluator/EvalContext/module runtime` | language intrinsic/runtime | `substrate/builtins_concurrency.go` |
| 92 | `_profile_counts` | `HAL.observe.profile_counts` | profile.counts (lib/profile/profile.ai:5) | `go:semantic measurement/runtime labels` | observation/runtime service | `substrate/builtins_profile.go` |
| 93 | `_profile_measure` | `HAL.observe.profile_measure` | profile.measure (lib/profile/profile.ai:10) | `go:semantic measurement/runtime labels` | observation/runtime service | `substrate/builtins_profile.go` |
| 94 | `_profile_experiment` | `HAL.observe.profile_experiment` | profile.experiment (lib/profile/profile.ai:16) | `go:semantic measurement/runtime labels` | observation/runtime service | `substrate/builtins_profile.go` |
| 95 | `_store_new` | `HAL.store.new` | store.new (lib/store/store.ai:5) | `go:value.Store` | language/value capability | `substrate/builtins_store.go` |
| 96 | `_store_get` | `HAL.store.get` | store.get (lib/store/store.ai:9) | `go:value.Store` | language/value capability | `substrate/builtins_store.go` |
| 97 | `_store_set` | `HAL.store.set` | store.set (lib/store/store.ai:13) | `go:value.Store` | language/value capability | `substrate/builtins_store.go` |
| 98 | `_store_length` | `HAL.store.length` | store.length (lib/store/store.ai:17) | `go:value.Store` | language/value capability | `substrate/builtins_store.go` |
| 99 | `_bits_and` | `HAL.bits.and` | bits.bit_and (lib/bits/bits.ai:5) | `go:math/big + value.Number` | language/value primitive | `substrate/builtins_bits.go` |
| 100 | `_bits_or` | `HAL.bits.or` | bits.bit_or (lib/bits/bits.ai:9) | `go:math/big + value.Number` | language/value primitive | `substrate/builtins_bits.go` |
| 101 | `_bits_xor` | `HAL.bits.xor` | bits.bit_xor (lib/bits/bits.ai:13) | `go:math/big + value.Number` | language/value primitive | `substrate/builtins_bits.go` |
| 102 | `_bits_not` | `HAL.bits.not` | bits.bit_not (lib/bits/bits.ai:17) | `go:math/big + value.Number` | language/value primitive | `substrate/builtins_bits.go` |
| 103 | `_bits_shl` | `HAL.bits.shl` | bits.shl (lib/bits/bits.ai:21); bits.mask (lib/bits/bits.ai:30) | `go:math/big + value.Number` | language/value primitive | `substrate/builtins_bits.go` |
| 104 | `_bits_shr` | `HAL.bits.shr` | bits.shr (lib/bits/bits.ai:25) | `go:math/big + value.Number` | language/value primitive | `substrate/builtins_bits.go` |
| 105 | `_quit` | `HAL.repl.quit` | quit (engine/runtime/prelude/prelude.ai:129) | `go:REPL/help/session state` | tooling/session service | `substrate/builtins_repl.go` |
| 106 | `_reset` | `HAL.repl.reset` | reset (engine/runtime/prelude/prelude.ai:133) | `go:REPL/help/session state` | tooling/session service | `substrate/builtins_repl.go` |
| 107 | `_delete` | `HAL.repl.delete` | delete (engine/runtime/prelude/prelude.ai:137) | `go:REPL/help/session state` | tooling/session service | `substrate/builtins_repl.go` |
| 108 | `_help` | `HAL.repl.help` | help (engine/runtime/prelude/prelude.ai:154); help (engine/runtime/prelude/prelude.ai:156) | `go:REPL/help/session state` | tooling/session service | `substrate/builtins_repl.go` |
| 109 | `_doc` | `HAL.repl.doc` | doc (engine/runtime/prelude/prelude.ai:161) | `go:REPL/help/session state` | tooling/session service | `substrate/builtins_repl.go` |
| 110 | `_test_equal` | `HAL.test.equal` | test.equal (lib/test/test.ai:9) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 111 | `_test_not_equal` | `HAL.test.not_equal` | test.not_equal (lib/test/test.ai:13) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 112 | `_test_true` | `HAL.test.true` | test.is_true (lib/test/test.ai:17) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 113 | `_test_false` | `HAL.test.false` | test.is_false (lib/test/test.ai:21) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 114 | `_test_error` | `HAL.test.error` | test.error (lib/test/test.ai:25) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 115 | `_test_not_error` | `HAL.test.not_error` | test.not_error (lib/test/test.ai:29) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 116 | `_test_run` | `HAL.test.run` | test.run (lib/test/test.ai:33) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |
| 117 | `_test_faults` | `HAL.test.faults` | test.faults (lib/test/test.ai:37) | `go:test harness/evaluator` | tooling/test service | `substrate/builtins_test_framework.go` |

## Registration groups

- IO: **2**
- List: **7**
- Type: **6**
- Math: **9**
- Bytes: **7**
- Time: **2**
- Host program environment: **3**
- Canvas: **17**
- Convert: **6**
- String (Unicode case conversion): **5**
- Regex: **6**
- File I/O: **12**
- Intrinsics - these use evaluation context: **9**
- Semantic work profiling: **3**
- Explicit mutable indexed storage: **4**
- Bit operations over non-negative integral Aiki numbers: **6**
- REPL: **5**
- Test framework: **8**
