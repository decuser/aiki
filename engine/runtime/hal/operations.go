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
	"_file_permissions": optionalPathOperation("HAL.file.permissions", "_file_permissions", "file.permissions", binding("file.permissions", "lib/file/file.ai")),
	"_file_chmod":       optionalPathOperation("HAL.file.chmod", "_file_chmod", "file.chmod", binding("file.chmod", "lib/file/file.ai")),
	"_time_now":         hostOperation("HAL.time.now", "_time_now", "runtime.clock", "nonblocking", "fault", bindings(binding("time.now", "lib/time/time.ai"))),
	"_system_cwd":       hostOperation("HAL.system.cwd", "_system_cwd", "runtime.process", "nonblocking", "fault", bindings(binding("system.cwd", "lib/system/system.ai"))),
	"_system_chdir":     hostOperation("HAL.system.chdir", "_system_chdir", "runtime.process", "may block", "fault or shaped :io result", bindings(binding("system.chdir", "lib/system/system.ai"))),
	"_path_separator":   hostOperation("HAL.path.separator", "_path_separator", "runtime.platform", "nonblocking", "fault", bindings(binding("path.separator", "lib/path/path.ai"))),
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
