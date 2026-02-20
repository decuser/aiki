// Package substrate implements the Go runtime substrate for HAL.
package substrate

import (
	"aiki/engine/semantics/value"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// GoRuntime implements hal.RuntimeContract using Go primitives.
type GoRuntime struct {
	registry map[string]func([]value.Value) (value.Value, error)
	mu       sync.RWMutex
}

// NewGoRuntime creates a new Go runtime substrate.
func NewGoRuntime() *GoRuntime {
	rt := &GoRuntime{
		registry: make(map[string]func([]value.Value) (value.Value, error)),
	}
	rt.registerDefaults()
	return rt
}

// Execute calls a registered primitive function.
func (g *GoRuntime) Execute(name string, args []value.Value) (value.Value, error) {
	g.mu.RLock()
	fn, ok := g.registry[name]
	g.mu.RUnlock()

	if !ok {
		return value.NullValue(), fmt.Errorf("native function %s not found in substrate", name)
	}

	return fn(args)
}

// HasBuiltin checks if a name is registered.
func (g *GoRuntime) HasBuiltin(name string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.registry[name]
	return ok
}

// Register adds a primitive function to the registry.
func (g *GoRuntime) Register(name string, fn func([]value.Value) (value.Value, error)) {
	g.mu.Lock()
	g.registry[name] = fn
	g.mu.Unlock()
}

// Spawn launches a goroutine.
func (g *GoRuntime) Spawn(task func()) {
	go task()
}

// MakeChannel creates a new channel value.
func (g *GoRuntime) MakeChannel() value.Value {
	ch := make(chan value.Value)
	return value.Value{Type: value.Channel, Data: ch}
}

// Send sends a value on a channel.
func (g *GoRuntime) Send(ch value.Value, val value.Value) error {
	if ch.Type != value.Channel {
		return fmt.Errorf("send: expected channel, got %s", value.TypeName(ch.Type))
	}
	c, ok := ch.Data.(chan value.Value)
	if !ok {
		return fmt.Errorf("send: invalid channel data")
	}
	c <- val
	return nil
}

// Recv receives a value from a channel.
func (g *GoRuntime) Recv(ch value.Value) (value.Value, error) {
	if ch.Type != value.Channel {
		return value.NullValue(), fmt.Errorf("recv: expected channel, got %s", value.TypeName(ch.Type))
	}
	c, ok := ch.Data.(chan value.Value)
	if !ok {
		return value.NullValue(), fmt.Errorf("recv: invalid channel data")
	}
	val := <-c
	return val, nil
}

// ReadFile reads a file from the filesystem.
func (g *GoRuntime) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ResolvePath resolves a module name to a file path.
func (g *GoRuntime) ResolvePath(name string, relativeTo string) (string, error) {
	// Try relative to current file
	if relativeTo != "" {
		dir := filepath.Dir(relativeTo)
		candidate := filepath.Join(dir, name+".ai")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Try as-is with .ai extension
	candidate := name + ".ai"
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// Try without extension
	if _, err := os.Stat(name); err == nil {
		return name, nil
	}

	return "", fmt.Errorf("module not found: %s", name)
}

// LogError logs an error (from spawned goroutines).
func (g *GoRuntime) LogError(err error) {
	fmt.Fprintf(os.Stderr, "spawn error: %v\n", err)
}

// registerDefaults registers the default primitive functions.
func (g *GoRuntime) registerDefaults() {
	// Print - writes arguments contiguously with no implicit separators
	g.Register("print", func(args []value.Value) (value.Value, error) {
		for _, a := range args {
			fmt.Print(a.Inspect())
		}
		return value.NullValue(), nil
	})

	// Println - print with newline
	g.Register("println", func(args []value.Value) (value.Value, error) {
		for i, a := range args {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(a.Inspect())
		}
		fmt.Println()
		return value.NullValue(), nil
	})

	// Type - returns the type of a value as a symbol
	g.Register("type", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("type: want 1 argument")
		}
		return value.NewSymbol(value.TypeName(args[0].Type)), nil
	})

	// Length - returns length of list or string
	g.Register("length", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("length: want 1 argument")
		}
		switch args[0].Type {
		case value.List:
			if list, ok := args[0].Data.([]value.Value); ok {
				return value.NewNumber(float64(len(list))), nil
			}
		case value.String:
			if s, ok := args[0].Data.(string); ok {
				return value.NewNumber(float64(len([]rune(s)))), nil
			}
		}
		return value.NullValue(), fmt.Errorf("length: invalid type %s", value.TypeName(args[0].Type))
	})

	// First - returns first element of list
	g.Register("first", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("first: want 1 argument")
		}
		if args[0].Type != value.List {
			return value.NullValue(), fmt.Errorf("first: expected list")
		}
		list, ok := args[0].Data.([]value.Value)
		if !ok || len(list) == 0 {
			return value.NullValue(), fmt.Errorf("first: empty list")
		}
		return list[0], nil
	})

	// Rest - returns all but first element
	g.Register("rest", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("rest: want 1 argument")
		}
		if args[0].Type != value.List {
			return value.NullValue(), fmt.Errorf("rest: expected list")
		}
		list, ok := args[0].Data.([]value.Value)
		if !ok || len(list) == 0 {
			return value.NewList(nil), nil
		}
		return value.NewList(list[1:]), nil
	})

	// Cons - prepends element to list
	g.Register("cons", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("cons: want 2 arguments")
		}
		if args[1].Type != value.List {
			return value.NullValue(), fmt.Errorf("cons: second argument must be list")
		}
		list, _ := args[1].Data.([]value.Value)
		newList := make([]value.Value, 0, len(list)+1)
		newList = append(newList, args[0])
		newList = append(newList, list...)
		return value.NewList(newList), nil
	})

	// Append - appends element to list
	g.Register("append", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("append: want 2 arguments")
		}
		if args[0].Type != value.List {
			return value.NullValue(), fmt.Errorf("append: first argument must be list")
		}
		list, _ := args[0].Data.([]value.Value)
		newList := make([]value.Value, len(list), len(list)+1)
		copy(newList, list)
		newList = append(newList, args[1])
		return value.NewList(newList), nil
	})

	// Equal - equality comparison
	g.Register("equal", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("equal: want 2 arguments")
		}
		return value.Value{Type: value.Boolean, Data: valuesEqual(args[0], args[1])}, nil
	})

	// Channel primitives
	g.Register("channel", func(args []value.Value) (value.Value, error) {
		ch := make(chan value.Value)
		return value.Value{Type: value.Channel, Data: ch}, nil
	})

	g.Register("send", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("send: want 2 arguments")
		}
		if args[0].Type != value.Channel {
			return value.NullValue(), fmt.Errorf("send: first argument must be channel")
		}
		ch, _ := args[0].Data.(chan value.Value)
		ch <- args[1]
		return value.True(), nil
	})

	g.Register("recv", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("recv: want 1 argument")
		}
		if args[0].Type != value.Channel {
			return value.NullValue(), fmt.Errorf("recv: argument must be channel")
		}
		ch, _ := args[0].Data.(chan value.Value)
		return <-ch, nil
	})

	// String conversion
	g.Register("to_str", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("to_str: want 1 argument")
		}
		return value.NewString(args[0].Inspect()), nil
	})

	// Modulo - remainder after division
	g.Register("modulo", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("modulo: want 2 arguments")
		}
		if args[0].Type != value.Number || args[1].Type != value.Number {
			return value.NullValue(), fmt.Errorf("modulo: arguments must be numbers")
		}
		a := int(args[0].Data.(float64))
		b := int(args[1].Data.(float64))
		if b == 0 {
			return value.NullValue(), fmt.Errorf("modulo: division by zero")
		}
		return value.NewNumber(float64(a % b)), nil
	})

	// Ord - get character code from rune/string
	g.Register("ord", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("ord: want 1 argument")
		}
		switch args[0].Type {
		case value.Rune:
			r := args[0].Data.(rune)
			return value.NewNumber(float64(r)), nil
		case value.String:
			s := args[0].Data.(string)
			if len(s) == 0 {
				return value.NewNumber(0), nil
			}
			runes := []rune(s)
			return value.NewNumber(float64(runes[0])), nil
		default:
			return value.NullValue(), fmt.Errorf("ord: expected rune or string")
		}
	})

	// Shape - get shape of a value
	g.Register("shape", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("shape: want 1 argument")
		}
		// For shaped lists, return the shape symbol
		// For now, return a symbol of the type
		return value.NewSymbol(value.TypeName(args[0].Type)), nil
	})

	// REPL builtins
	g.Register("help", func(args []value.Value) (value.Value, error) {
		helpText := `Aiki

Primitives:
  first(list)         - first element
  rest(list)          - all but first
  length(list)        - length
  cons(val, list)     - prepend to list
  append(list, val)   - append to list
  type(val)           - type as symbol
  shape(val)          - shape name or :list
  equal(a, b)         - deep equality
  to_str(val)         - convert to string
  print(val...)       - output (no newline)
  println(val...)     - output with newline
  modulo(a, b)        - remainder
  ord(rune)           - character code

REPL:
  help()              - this message
  quit()              - exit REPL
`
		fmt.Print(helpText)
		return value.NullValue(), nil
	})

	g.Register("quit", func(args []value.Value) (value.Value, error) {
		return value.NewSymbol("exit"), nil
	})

	g.Register("exit", func(args []value.Value) (value.Value, error) {
		return value.NewSymbol("exit"), nil
	})

	g.Register("reset", func(args []value.Value) (value.Value, error) {
		return value.NewSymbol("reset"), nil
	})
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b value.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case value.Null:
		return true
	case value.Number:
		av, _ := a.AsNumber()
		bv, _ := b.AsNumber()
		return av == bv
	case value.Boolean:
		av, _ := a.AsBool()
		bv, _ := b.AsBool()
		return av == bv
	case value.String:
		av, _ := a.AsString()
		bv, _ := b.AsString()
		return av == bv
	case value.Symbol:
		av, _ := a.Data.(string)
		bv, _ := b.Data.(string)
		return av == bv
	case value.Rune:
		av, _ := a.Data.(rune)
		bv, _ := b.Data.(rune)
		return av == bv
	case value.List:
		al, _ := a.AsList()
		bl, _ := b.AsList()
		if len(al) != len(bl) {
			return false
		}
		for i := range al {
			if !valuesEqual(al[i], bl[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
