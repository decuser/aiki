# Turtle Patch

Extract this archive over your aiki repository root.

## Validation

After patching:

```bash
make validate
make runsamples
make rigorous
```

## Contents

- `engine/runtime/hal/substrate/builtins_canvas_accessors.go` - width/height builtins
- `engine/runtime/hal/substrate/builtins_canvas_accessors_test.go` - Go tests
- `lib/canvas/canvas.ai` - updated with width/height exports
- `lib/canvas/canvas.help` - updated help
- `lib/canvas/canvas.doc` - updated docs
- `lib/turtle/turtle.ai` - turtle implementation
- `lib/turtle/turtle.help` - turtle help
- `lib/turtle/turtle.doc` - turtle docs
- `test/behavior/turtle_smoke.ai` - smoke test
- `test/behavior/turtle_smoke.gold` - expected output
