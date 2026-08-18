package hal

var operationDefinitions = map[string]HostOperation{
	"_print": hostOperation("HAL.io.print", "_print", "runtime.io", "may block", "fault", bindings(
		binding("print", "engine/runtime/prelude/prelude.ai"),
		binding("println", "engine/runtime/prelude/prelude.ai"),
	)),
	"_read": hostOperation("HAL.io.read", "_read", "runtime.io", "may block", "fault or shaped :io/:end result", bindings(
		binding("read", "engine/runtime/prelude/prelude.ai"),
		binding("input", "engine/runtime/prelude/prelude.ai"),
	)),
	"_io_read": hostOperation("HAL.io.read_bytes", "_io_read", "runtime.io", "may block", "fault or shaped :io/:end result", bindings(
		binding("io.read", "lib/io/io.ai"),
	)),
	"_io_read_line": hostOperation("HAL.io.read_line", "_io_read_line", "runtime.io", "may block", "fault or shaped :io/:end result", bindings(
		binding("io.read_line", "lib/io/io.ai"),
	)),
	"_io_write": hostOperation("HAL.io.write", "_io_write", "runtime.io", "may block", "fault or shaped :io result", bindings(
		binding("io.write", "lib/io/io.ai"),
	)),
	"_io_close": {
		Identity: "HAL.io.close", Primitive: "_io_close", Authority: "HAL.io.close",
		Context: []string{"runtime.io", "runtime.resources"}, Effect: "state mutation",
		Blocking: "may block", Lifetime: "operates on resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :io result", SemanticObservation: "io.close",
		AikiBindings: bindings(binding("io.close", "lib/io/io.ai")),
	},
	"_sleep": hostOperation("HAL.time.sleep", "_sleep", "runtime.clock", "waits externally", "fault", bindings(
		binding("time.sleep", "lib/time/time.ai"),
	)),
	"_after": {
		Identity:            "HAL.time.after",
		Primitive:           "_after",
		Authority:           "HAL.time.after",
		Context:             []string{"runtime.clock"},
		Effect:              "async source",
		Blocking:            "nonblocking",
		Lifetime:            "returns resource",
		Optionality:         "constitutive",
		ErrorContract:       "fault",
		SemanticObservation: "time.after",
		AikiBindings:        bindings(binding("time.after", "lib/time/time.ai")),
	},
	"_system_args": hostOperation("HAL.system.args", "_system_args", "runtime.process", "nonblocking", "fault", bindings(
		binding("system.args", "lib/system/system.ai"),
	)),
	"_system_env": hostOperation("HAL.system.env", "_system_env", "runtime.process", "nonblocking", "fault or shaped :environment result", bindings(
		binding("system.env", "lib/system/system.ai"),
	)),
	"_system_environ": hostOperation("HAL.system.environ", "_system_environ", "runtime.process", "nonblocking", "fault", bindings(
		binding("system.environ", "lib/system/system.ai"),
	)),
	"_system_set_env": hostOperation("HAL.system.set_env", "_system_set_env", "runtime.process", "nonblocking", "fault or shaped :environment result", bindings(
		binding("system.set_env", "lib/system/system.ai"),
	)),
	"_system_unset_env": hostOperation("HAL.system.unset_env", "_system_unset_env", "runtime.process", "nonblocking", "fault or shaped :environment result", bindings(
		binding("system.unset_env", "lib/system/system.ai"),
	)),
	"_file_open":        resourceAcquisition("HAL.file.open", "_file_open", "file.open", binding("file.open", "lib/file/file.ai")),
	"_file_read_text":   resourceOperation("HAL.file.read_text", "_file_read_text", "file.read_text", binding("file.read_text", "lib/file/file.ai")),
	"_file_read_bytes":  resourceOperation("HAL.file.read_bytes", "_file_read_bytes", "file.read_bytes", binding("file.read_bytes", "lib/file/file.ai")),
	"_file_read_line":   resourceOperation("HAL.file.read_line", "_file_read_line", "file.read_line", binding("file.read_line", "lib/file/file.ai")),
	"_file_write_text":  resourceOperation("HAL.file.write_text", "_file_write_text", "file.write_text", binding("file.write_text", "lib/file/file.ai")),
	"_file_write_bytes": resourceOperation("HAL.file.write_bytes", "_file_write_bytes", "file.write_bytes", binding("file.write_bytes", "lib/file/file.ai")),
	"_file_close":       resourceOperation("HAL.file.close", "_file_close", "file.close", binding("file.close", "lib/file/file.ai")),
	"_file_exists":      pathOperation("HAL.file.exists", "_file_exists", "file.exists", binding("file.exists", "lib/file/file.ai")),
	"_file_delete":      pathOperation("HAL.file.delete", "_file_delete", "file.delete", binding("file.delete", "lib/file/file.ai")),
	"_file_list":        pathOperation("HAL.file.list", "_file_list", "file.list", binding("file.list", "lib/file/file.ai")),
	"_file_read_at":     resourceOperation("HAL.file.read_at", "_file_read_at", "file.read_at", binding("file.read_at", "lib/file/file.ai")),
	"_file_write_at":    resourceOperation("HAL.file.write_at", "_file_write_at", "file.write_at", binding("file.write_at", "lib/file/file.ai")),
	"_file_stat":        pathOperation("HAL.file.stat", "_file_stat", "file.stat", binding("file.stat", "lib/file/file.ai")),
	"_file_rename":      pathOperation("HAL.file.rename", "_file_rename", "file.rename", binding("file.rename", "lib/file/file.ai")),
	"_file_mkdir":       pathOperation("HAL.file.mkdir", "_file_mkdir", "file.mkdir", binding("file.mkdir", "lib/file/file.ai")),
	"_file_mkdir_all":   pathOperation("HAL.file.mkdir_all", "_file_mkdir_all", "file.mkdir_all", binding("file.mkdir_all", "lib/file/file.ai")),
	"_file_remove_all":  pathOperation("HAL.file.remove_all", "_file_remove_all", "file.remove_all", binding("file.remove_all", "lib/file/file.ai")),
	"_file_temp":        pathOperation("HAL.file.temp", "_file_temp", "file.temp", binding("file.temp", "lib/file/file.ai")),
	"_file_temp_dir":    pathOperation("HAL.file.temp_dir", "_file_temp_dir", "file.temp_dir", binding("file.temp_dir", "lib/file/file.ai")),
	"_file_copy":        pathOperation("HAL.file.copy", "_file_copy", "file.copy", binding("file.copy", "lib/file/file.ai")),
	"_file_size":        pathOperation("HAL.file.size", "_file_size", "file.size", binding("file.size", "lib/file/file.ai")),
	"_file_walk":        pathOperation("HAL.file.walk", "_file_walk", "file.walk", binding("file.walk", "lib/file/file.ai")),
	"_file_symlink":     optionalPathOperation("HAL.file.symlink", "_file_symlink", "file.symlink", binding("file.symlink", "lib/file/file.ai")),
	"_file_read_link":   optionalPathOperation("HAL.file.read_link", "_file_read_link", "file.read_link", binding("file.read_link", "lib/file/file.ai")),
	"_file_lock": {
		Identity:            "HAL.file.lock",
		Primitive:           "_file_lock",
		Authority:           "HAL.file.lock",
		Context:             []string{"runtime.filesystem", "runtime.resources"},
		Effect:              "resource acquisition",
		Blocking:            "may block",
		Lifetime:            "returns resource",
		Optionality:         "constitutive",
		ErrorContract:       "fault or shaped :lock result",
		SemanticObservation: "file.lock",
		AikiBindings:        bindings(binding("file.lock", "lib/file/file.ai")),
	},
	"_file_try_lock": {
		Identity:            "HAL.file.try_lock",
		Primitive:           "_file_try_lock",
		Authority:           "HAL.file.try_lock",
		Context:             []string{"runtime.filesystem", "runtime.resources"},
		Effect:              "resource acquisition",
		Blocking:            "nonblocking",
		Lifetime:            "returns resource",
		Optionality:         "constitutive",
		ErrorContract:       "fault, false, or shaped :lock result",
		SemanticObservation: "file.try_lock",
		AikiBindings:        bindings(binding("file.try_lock", "lib/file/file.ai")),
	},
	"_file_unlock": {
		Identity:            "HAL.file.unlock",
		Primitive:           "_file_unlock",
		Authority:           "HAL.file.unlock",
		Context:             []string{"runtime.resources"},
		Effect:              "state mutation",
		Blocking:            "nonblocking",
		Lifetime:            "operates on resource",
		Optionality:         "constitutive",
		ErrorContract:       "fault or shaped :lock result",
		SemanticObservation: "file.unlock",
		AikiBindings:        bindings(binding("file.unlock", "lib/file/file.ai")),
	},
	"_file_permissions": optionalPathOperation("HAL.file.permissions", "_file_permissions", "file.permissions", binding("file.permissions", "lib/file/file.ai")),
	"_file_chmod":       optionalPathOperation("HAL.file.chmod", "_file_chmod", "file.chmod", binding("file.chmod", "lib/file/file.ai")),
	"_seed": {
		Identity:            "HAL.random.seed",
		Primitive:           "_seed",
		Authority:           "HAL.random.seed",
		Context:             []string{"runtime.random"},
		Effect:              "state mutation",
		Blocking:            "nonblocking",
		Lifetime:            "runtime-owned state",
		Optionality:         "constitutive",
		ErrorContract:       "fault",
		SemanticObservation: "random.seed",
		AikiBindings:        bindings(binding("random.seed", "lib/random/random.ai")),
	},
	"_random": {
		Identity:            "HAL.random.below",
		Primitive:           "_random",
		Authority:           "HAL.random.below",
		Context:             []string{"runtime.random"},
		Effect:              "state mutation",
		Blocking:            "nonblocking",
		Lifetime:            "runtime-owned state",
		Optionality:         "constitutive",
		ErrorContract:       "fault",
		SemanticObservation: "random.random",
		AikiBindings:        bindings(binding("random.random", "lib/random/random.ai")),
	},
	"_time_now":       hostOperation("HAL.time.now", "_time_now", "runtime.clock", "nonblocking", "fault", bindings(binding("time.now", "lib/time/time.ai"))),
	"_system_cwd":     hostOperation("HAL.system.cwd", "_system_cwd", "runtime.process", "nonblocking", "fault", bindings(binding("system.cwd", "lib/system/system.ai"))),
	"_system_chdir":   hostOperation("HAL.system.chdir", "_system_chdir", "runtime.process", "may block", "fault or shaped :io result", bindings(binding("system.chdir", "lib/system/system.ai"))),
	"_path_separator": hostOperation("HAL.path.separator", "_path_separator", "runtime.platform", "nonblocking", "fault", bindings(binding("path.separator", "lib/path/path.ai"))),
	"_canvas": {
		Identity:            "HAL.canvas.open",
		Primitive:           "_canvas",
		Authority:           "HAL.canvas.open",
		Context:             []string{"runtime.resources", "runtime.graphics"},
		Effect:              "resource acquisition",
		Blocking:            "waits externally",
		Lifetime:            "returns resource",
		Optionality:         "optional",
		ErrorContract:       "fault or shaped :canvas result",
		SemanticObservation: "canvas.canvas",
		AikiBindings:        bindings(binding("canvas.canvas", "lib/canvas/canvas.ai")),
	},
	"_canvas_command": {
		Identity:            "HAL.canvas.command",
		Primitive:           "_canvas_command",
		Authority:           "HAL.canvas.command",
		Context:             []string{"runtime.resources", "runtime.graphics"},
		Effect:              "bounded call",
		Blocking:            "may block",
		Lifetime:            "operates on resource",
		Optionality:         "optional",
		ErrorContract:       "fault or shaped :canvas result",
		SemanticObservation: "canvas.command",
		AikiBindings: bindings(
			binding("canvas.dot", "lib/canvas/canvas.ai"),
			binding("canvas.line", "lib/canvas/canvas.ai"),
			binding("canvas.rect", "lib/canvas/canvas.ai"),
			binding("canvas.fill_rect", "lib/canvas/canvas.ai"),
			binding("canvas.circle", "lib/canvas/canvas.ai"),
			binding("canvas.fill_circle", "lib/canvas/canvas.ai"),
			binding("canvas.arc", "lib/canvas/canvas.ai"),
			binding("canvas.clear", "lib/canvas/canvas.ai"),
			binding("canvas.set_bg", "lib/canvas/canvas.ai"),
			binding("canvas.set_fg", "lib/canvas/canvas.ai"),
			binding("canvas.pen_size", "lib/canvas/canvas.ai"),
		),
	},
	"_destroy":       canvasResourceOperation("HAL.canvas.close", "_destroy", "canvas.destroy", binding("canvas.destroy", "lib/canvas/canvas.ai")),
	"_canvas_width":  canvasResourceOperation("HAL.canvas.width", "_canvas_width", "canvas.width", binding("canvas.width", "lib/canvas/canvas.ai")),
	"_canvas_height": canvasResourceOperation("HAL.canvas.height", "_canvas_height", "canvas.height", binding("canvas.height", "lib/canvas/canvas.ai")),
	"_canvas_alive":  canvasResourceOperation("HAL.canvas.alive", "_canvas_alive", "canvas.alive", binding("canvas.alive", "lib/canvas/canvas.ai")),
	"_process_start": {
		Identity: "HAL.process.start", Primitive: "_process_start", Authority: "HAL.process.start",
		Context: []string{"runtime.process", "runtime.io", "runtime.resources"}, Effect: "resource acquisition",
		Blocking: "may block", Lifetime: "returns resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :process result", SemanticObservation: "process.start",
		AikiBindings: bindings(binding("process.start", "lib/process/process.ai")),
	},
	"_process_stdin":  processResourceOperation("HAL.process.stdin", "_process_stdin", "process.stdin"),
	"_process_stdout": processResourceOperation("HAL.process.stdout", "_process_stdout", "process.stdout"),
	"_process_stderr": processResourceOperation("HAL.process.stderr", "_process_stderr", "process.stderr"),
	"_process_wait": {
		Identity: "HAL.process.wait", Primitive: "_process_wait", Authority: "HAL.process.wait",
		Context: []string{"runtime.process", "runtime.resources"}, Effect: "bounded call",
		Blocking: "waits externally", Lifetime: "operates on resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :process result", SemanticObservation: "process.wait",
		AikiBindings: bindings(binding("process.wait", "lib/process/process.ai")),
	},
	"_process_terminate": {
		Identity: "HAL.process.terminate", Primitive: "_process_terminate", Authority: "HAL.process.terminate",
		Context: []string{"runtime.process", "runtime.resources"}, Effect: "state mutation",
		Blocking: "nonblocking", Lifetime: "operates on resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :process result", SemanticObservation: "process.terminate",
		AikiBindings: bindings(binding("process.terminate", "lib/process/process.ai")),
	},
	"_signal_watch": {
		Identity: "HAL.signal.watch", Primitive: "_signal_watch", Authority: "HAL.signal.watch",
		Context: []string{"runtime.signals", "runtime.resources"}, Effect: "resource acquisition",
		Blocking: "nonblocking", Lifetime: "returns resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :signal/:unsupported result", SemanticObservation: "signal.watch",
		AikiBindings: bindings(binding("signal.watch", "lib/signal/signal.ai")),
	},
	"_signal_stop": {
		Identity: "HAL.signal.stop", Primitive: "_signal_stop", Authority: "HAL.signal.stop",
		Context: []string{"runtime.signals", "runtime.resources"}, Effect: "state mutation",
		Blocking: "nonblocking", Lifetime: "operates on resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :signal result", SemanticObservation: "signal.stop",
		AikiBindings: bindings(binding("signal.stop", "lib/signal/signal.ai")),
	},
	"_signal_send": {
		Identity: "HAL.signal.send", Primitive: "_signal_send", Authority: "HAL.signal.send",
		Context: []string{"runtime.signals", "runtime.process", "runtime.resources"}, Effect: "state mutation",
		Blocking: "nonblocking", Lifetime: "operates on resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :signal/:unsupported result", SemanticObservation: "signal.send",
		AikiBindings: bindings(binding("signal.send", "lib/signal/signal.ai")),
	},
	"_net_connect":  networkOperation("HAL.net.connect", "_net_connect", "net.connect", "resource acquisition", "may block", "returns resource"),
	"_net_listen":   networkOperation("HAL.net.listen", "_net_listen", "net.listen", "resource acquisition", "may block", "returns resource"),
	"_net_accept":   networkOperation("HAL.net.accept", "_net_accept", "net.accept", "resource acquisition", "waits externally", "returns resource"),
	"_net_local":    networkOperation("HAL.net.local", "_net_local", "net.local", "bounded call", "nonblocking", "operates on resource"),
	"_net_remote":   networkOperation("HAL.net.remote", "_net_remote", "net.remote", "bounded call", "nonblocking", "operates on resource"),
	"_net_close":    networkOperation("HAL.net.close", "_net_close", "net.close", "state mutation", "nonblocking", "operates on resource"),
	"_net_udp_bind": networkOperation("HAL.net.udp_bind", "_net_udp_bind", "net.udp_bind", "resource acquisition", "may block", "returns resource"),
	"_net_udp_send": networkOperation("HAL.net.udp_send", "_net_udp_send", "net.udp_send", "bounded call", "may block", "operates on resource"),
	"_net_udp_recv": networkOperation("HAL.net.udp_recv", "_net_udp_recv", "net.udp_recv", "bounded call", "waits externally", "operates on resource"),
	"_term_is": {
		Identity: "HAL.term.is", Primitive: "_term_is", Authority: "HAL.term.is",
		Context: []string{"runtime.terminal", "runtime.io"}, Effect: "bounded call",
		Blocking: "nonblocking", Lifetime: "stateless", Optionality: "constitutive",
		ErrorContract: "fault or shaped :terminal result", SemanticObservation: "term.is",
		AikiBindings: bindings(binding("term.is", "lib/term/term.ai")),
	},
	"_term_size": {
		Identity: "HAL.term.size", Primitive: "_term_size", Authority: "HAL.term.size",
		Context: []string{"runtime.terminal", "runtime.io"}, Effect: "bounded call",
		Blocking: "nonblocking", Lifetime: "stateless", Optionality: "constitutive",
		ErrorContract: "fault or shaped :terminal result", SemanticObservation: "term.size",
		AikiBindings: bindings(binding("term.size", "lib/term/term.ai")),
	},
	"_term_raw": {
		Identity: "HAL.term.raw", Primitive: "_term_raw", Authority: "HAL.term.raw",
		Context: []string{"runtime.terminal", "runtime.io", "runtime.resources"}, Effect: "resource acquisition",
		Blocking: "nonblocking", Lifetime: "returns resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :terminal result", SemanticObservation: "term.raw",
		AikiBindings: bindings(binding("term.raw", "lib/term/term.ai")),
	},
	"_term_restore": {
		Identity: "HAL.term.restore", Primitive: "_term_restore", Authority: "HAL.term.restore",
		Context: []string{"runtime.terminal", "runtime.resources"}, Effect: "state mutation",
		Blocking: "nonblocking", Lifetime: "operates on resource", Optionality: "constitutive",
		ErrorContract: "fault or shaped :terminal result", SemanticObservation: "term.restore",
		AikiBindings: bindings(binding("term.restore", "lib/term/term.ai")),
	},
	"_system_exec": {
		Identity:            "HAL.process.exec",
		Primitive:           "_system_exec",
		Authority:           "HAL.process.exec",
		Context:             []string{"runtime.process"},
		Effect:              "bounded call",
		Blocking:            "waits externally",
		Lifetime:            "stateless",
		Optionality:         "constitutive",
		ErrorContract:       "fault or shaped :process result",
		SemanticObservation: "system.exec",
		AikiBindings:        bindings(binding("system.exec", "lib/system/system.ai")),
	},
}

func binding(name, source string) AikiBinding {
	return AikiBinding{Name: name, Source: source}
}

func bindings(bs ...AikiBinding) []AikiBinding { return bs }

func hostOperation(identity, primitive, context, blocking, errors string, aiki []AikiBinding) HostOperation {
	observation := ""
	if len(aiki) > 0 {
		observation = aiki[0].Name
	}
	return HostOperation{
		Identity:            identity,
		Primitive:           primitive,
		Authority:           identity,
		Context:             []string{context},
		Effect:              "bounded call",
		Blocking:            blocking,
		Lifetime:            "stateless",
		Optionality:         "constitutive",
		ErrorContract:       errors,
		SemanticObservation: observation,
		AikiBindings:        aiki,
	}
}

func fileOperation(identity, primitive, observation string, b AikiBinding) HostOperation {
	op := hostOperation(identity, primitive, "runtime.filesystem", "may block", "fault or shaped :io result", bindings(b))
	op.SemanticObservation = observation
	return op
}

func pathOperation(identity, primitive, observation string, b AikiBinding) HostOperation {
	return fileOperation(identity, primitive, observation, b)
}

func optionalPathOperation(identity, primitive, observation string, b AikiBinding) HostOperation {
	op := fileOperation(identity, primitive, observation, b)
	op.Optionality = "optional"
	return op
}

func resourceAcquisition(identity, primitive, observation string, b AikiBinding) HostOperation {
	op := fileOperation(identity, primitive, observation, b)
	op.Effect = "resource acquisition"
	op.Lifetime = "returns resource"
	return op
}

func resourceOperation(identity, primitive, observation string, b AikiBinding) HostOperation {
	op := fileOperation(identity, primitive, observation, b)
	op.Lifetime = "operates on resource"
	return op
}

func canvasResourceOperation(identity, primitive, observation string, b AikiBinding) HostOperation {
	op := hostOperation(identity, primitive, "runtime.resources", "nonblocking", "fault or shaped :canvas result", bindings(b))
	op.Lifetime = "operates on resource"
	op.Optionality = "optional"
	op.SemanticObservation = observation
	return op
}

func processResourceOperation(identity, primitive, observation string) HostOperation {
	op := hostOperation(identity, primitive, "runtime.process", "nonblocking", "fault or shaped :process result", bindings(binding(observation, "lib/process/process.ai")))
	op.Context = []string{"runtime.process", "runtime.resources"}
	op.Lifetime = "operates on resource"
	op.SemanticObservation = observation
	return op
}

func networkOperation(identity, primitive, observation, effect, blocking, lifetime string) HostOperation {
	return HostOperation{
		Identity: identity, Primitive: primitive, Authority: identity,
		Context: []string{"runtime.network", "runtime.io", "runtime.resources"}, Effect: effect,
		Blocking: blocking, Lifetime: lifetime, Optionality: "constitutive",
		ErrorContract: "fault or shaped :network result", SemanticObservation: observation,
		AikiBindings: bindings(binding(observation, "lib/net/net.ai")),
	}
}

// OperationDefinition returns the architectural contract for a substrate primitive.
// Substrate provenance is intentionally supplied by the substrate binding layer.
func OperationDefinition(primitive string) (HostOperation, bool) {
	op, ok := operationDefinitions[primitive]
	if !ok {
		return HostOperation{}, false
	}
	op.Context = append([]string(nil), op.Context...)
	op.AikiBindings = append([]AikiBinding(nil), op.AikiBindings...)
	return op, true
}

// OperationDefinitions returns an independent copy of the canonical HAL operation set.
func OperationDefinitions() map[string]HostOperation {
	out := make(map[string]HostOperation, len(operationDefinitions))
	for primitive := range operationDefinitions {
		op, _ := OperationDefinition(primitive)
		out[primitive] = op
	}
	return out
}
