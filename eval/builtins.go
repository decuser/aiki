package eval

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strings"

	"aiki/value"
)

var builtins = map[string]*value.Builtin{
	// List operations
	"first": {
		Name: "first",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("first: want 1 argument, got %d", len(args))
			}
			switch a := args[0].(type) {
			case *value.List:
				if len(a.Elements) == 0 {
					return value.NewError("first: empty list")
				}
				return a.Elements[0]
			case *value.String:
				if len(a.Value) == 0 {
					return value.NewError("first: empty string")
				}
				return &value.Rune{Value: []rune(a.Value)[0]}
			default:
				return value.NewError("first: expected list or string")
			}
		},
	},
	"rest": {
		Name: "rest",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("rest: want 1 argument, got %d", len(args))
			}
			switch a := args[0].(type) {
			case *value.List:
				if len(a.Elements) == 0 {
					return &value.List{Elements: []value.Value{}}
				}
				return &value.List{Elements: a.Elements[1:]}
			case *value.String:
				if len(a.Value) == 0 {
					return &value.String{Value: ""}
				}
				return &value.String{Value: string([]rune(a.Value)[1:])}
			default:
				return value.NewError("rest: expected list or string")
			}
		},
	},
	"len": {
		Name: "len",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("len: want 1 argument, got %d", len(args))
			}
			switch a := args[0].(type) {
			case *value.List:
				return value.NewNumber(int64(len(a.Elements)), 1)
			case *value.String:
				return value.NewNumber(int64(len([]rune(a.Value))), 1)
			case *value.Bytes:
				return value.NewNumber(int64(len(a.Value)), 1)
			default:
				return value.NewError("len: expected list, string, or bytes")
			}
		},
	},
	"prepend": {
		Name: "prepend",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return value.NewError("prepend: want 2 arguments, got %d", len(args))
			}
			list, ok := args[0].(*value.List)
			if !ok {
				return value.NewError("prepend: first argument must be list")
			}
			newElements := make([]value.Value, len(list.Elements)+1)
			newElements[0] = args[1]
			copy(newElements[1:], list.Elements)
			return &value.List{Elements: newElements}
		},
	},
	"append": {
		Name: "append",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return value.NewError("append: want 2 arguments, got %d", len(args))
			}
			list, ok := args[0].(*value.List)
			if !ok {
				return value.NewError("append: first argument must be list")
			}
			newElements := make([]value.Value, len(list.Elements)+1)
			copy(newElements, list.Elements)
			newElements[len(list.Elements)] = args[1]
			return &value.List{Elements: newElements}
		},
	},
	"nth": {
		Name: "nth",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return value.NewError("nth: want 2 arguments, got %d", len(args))
			}
			list, ok := args[0].(*value.List)
			if !ok {
				return value.NewError("nth: first argument must be list")
			}
			idx, ok := args[1].(*value.Number)
			if !ok || !idx.Value.IsInt() {
				return value.NewError("nth: second argument must be integer")
			}
			i := int(idx.Value.Num().Int64())
			if i < 0 || i >= len(list.Elements) {
				return value.NewError("nth: index out of bounds: %d", i)
			}
			return list.Elements[i]
		},
	},

	// Type operations
	"type": {
		Name: "type",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("type: want 1 argument, got %d", len(args))
			}
			return &value.Symbol{Value: string(args[0].Type())}
		},
	},
	"inspect": {
		Name: "inspect",
		Fn: func(args ...value.Value) value.Value {
			if len(args) < 1 {
				return value.NewError("inspect: want at least 1 argument")
			}
			return &value.String{Value: args[0].Inspect()}
		},
	},
	"shape": {
		Name: "shape",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("shape: want 1 argument, got %d", len(args))
			}
			list, ok := args[0].(*value.List)
			if !ok {
				return &value.Symbol{Value: "list"}
			}
			if list.Shape == "" {
				return &value.Symbol{Value: "list"}
			}
			return &value.Symbol{Value: list.Shape}
		},
	},

	// Compare
	"equal": {
		Name: "equal",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return value.NewError("equal: want 2 arguments, got %d", len(args))
			}
			return nativeBoolToBoolean(deepEqual(args[0], args[1]))
		},
	},

	// Convert
	"tostr": {
		Name: "tostr",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("tostr: want 1 argument, got %d", len(args))
			}
			return &value.String{Value: args[0].Inspect()}
		},
	},
	"tonum": {
		Name: "tonum",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("tonum: want 1 argument, got %d", len(args))
			}
			s, ok := args[0].(*value.String)
			if !ok {
				return makeError("tonum: expected string")
			}
			n, err := value.NewNumberFromString(s.Value)
			if err != nil {
				return makeError("invalid number: " + s.Value)
			}
			return makeOk(n)
		},
	},
	"todecimal": {
		Name: "todecimal",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return value.NewError("todecimal: want 2 arguments, got %d", len(args))
			}
			n, ok := args[0].(*value.Number)
			if !ok {
				return value.NewError("todecimal: first argument must be number")
			}
			places, ok := args[1].(*value.Number)
			if !ok || !places.Value.IsInt() {
				return value.NewError("todecimal: second argument must be integer")
			}
			p := int(places.Value.Num().Int64())
			f, _ := n.Value.Float64()
			format := fmt.Sprintf("%%.%df", p)
			return &value.String{Value: fmt.Sprintf(format, f)}
		},
	},

	// I/O
	"print": {
		Name: "print",
		Fn: func(args ...value.Value) value.Value {
			parts := make([]string, len(args))
			for i, a := range args {
				switch v := a.(type) {
				case *value.String:
					parts[i] = v.Value
				default:
					parts[i] = a.Inspect()
				}
			}
			fmt.Println(strings.Join(parts, " "))
			return value.NULL
		},
	},
	"read": {
		Name: "read",
		Fn: func(args ...value.Value) value.Value {
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if line == "" {
						return &value.List{
							Elements: []value.Value{},
							Shape:    "end",
						}
					}
					// Return partial line at EOF
					return &value.String{Value: line}
				}
				return makeError("read: " + err.Error())
			}
			// Strip trailing newline
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			return &value.String{Value: line}
		},
	},

	// Math
	"sqrt": {
		Name: "sqrt",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("sqrt: want 1 argument, got %d", len(args))
			}
			n, ok := args[0].(*value.Number)
			if !ok {
				return value.NewError("sqrt: expected number")
			}
			f, _ := n.Value.Float64()
			if f < 0 {
				return value.NewError("sqrt: negative number")
			}
			r := new(big.Rat).SetFloat64(math.Sqrt(f))
			return &value.Number{Value: r}
		},
	},
	"sin": {
		Name: "sin",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("sin: want 1 argument, got %d", len(args))
			}
			n, ok := args[0].(*value.Number)
			if !ok {
				return value.NewError("sin: expected number")
			}
			f, _ := n.Value.Float64()
			r := new(big.Rat).SetFloat64(math.Sin(f))
			return &value.Number{Value: r}
		},
	},
	"cos": {
		Name: "cos",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("cos: want 1 argument, got %d", len(args))
			}
			n, ok := args[0].(*value.Number)
			if !ok {
				return value.NewError("cos: expected number")
			}
			f, _ := n.Value.Float64()
			r := new(big.Rat).SetFloat64(math.Cos(f))
			return &value.Number{Value: r}
		},
	},
	"random": {
		Name: "random",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("random: want 1 argument, got %d", len(args))
			}
			n, ok := args[0].(*value.Number)
			if !ok || !n.Value.IsInt() {
				return value.NewError("random: expected integer")
			}
			max := n.Value.Num().Int64()
			if max <= 0 {
				return value.NewError("random: max must be positive")
			}
			return value.NewNumber(rand.Int63n(max), 1)
		},
	},

	// System
	"help": {
		Name: "help",
		Fn: func(args ...value.Value) value.Value {
			help := `Aiki v0.2.0

Primitives:
  first(list)         - first element
  rest(list)          - all but first
  len(list)           - length
  prepend(list val)   - add to front
  append(list val)    - add to end
  type(val)           - type as symbol
  inspect(val)        - string representation
  shape(val)          - shape name or :list
  equal(a b)          - deep equality
  tostr(val)          - convert to string
  tonum(str)          - parse number
  todecimal(n places) - format decimal
  print(val...)       - output
  read()              - input line
  sqrt(n)             - square root
  sin(n)              - sine
  cos(n)              - cosine
  random(n)           - random 0 to n-1

Type help(name) for details.`
			return &value.String{Value: help}
		},
	},
	"quit": {
		Name: "quit",
		Fn: func(args ...value.Value) value.Value {
			fmt.Println("Goodbye!")
			return value.NULL
		},
	},

	// Internal - used by prelude
	"_hash_code": {
		Name: "_hash_code",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return value.NewError("_hash_code: want 1 argument, got %d", len(args))
			}
			return value.NewNumber(hashValue(args[0]), 1)
		},
	},
"open": {
		Name: "open",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return makeError("open: want 1 argument")
			}
			path, ok := args[0].(*value.String)
			if !ok {
				return makeError("open: expected string path")
			}
			f, err := os.Open(path.Value)
			if err != nil {
				return makeError("open: " + err.Error())
			}
			return &value.Handle{File: f, Path: path.Value}
		},
	},
	"create": {
		Name: "create",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return makeError("create: want 1 argument")
			}
			path, ok := args[0].(*value.String)
			if !ok {
				return makeError("create: expected string path")
			}
			f, err := os.Create(path.Value)
			if err != nil {
				return makeError("create: " + err.Error())
			}
			return &value.Handle{File: f, Path: path.Value}
		},
	},
	"fread": {
		Name: "fread",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return makeError("fread: want 1 argument")
			}
			h, ok := args[0].(*value.Handle)
			if !ok {
				return makeError("fread: expected handle")
			}
			buf := make([]byte, 4096)
			n, err := h.File.Read(buf)
			if err != nil {
				if err == io.EOF {
					return &value.List{
						Elements: []value.Value{},
						Shape:    "end",
					}
				}
				return makeError("fread: " + err.Error())
			}
			return &value.String{Value: string(buf[:n])}
		},
	},
	"fwrite": {
		Name: "fwrite",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 2 {
				return makeError("fwrite: want 2 arguments")
			}
			h, ok := args[0].(*value.Handle)
			if !ok {
				return makeError("fwrite: expected handle")
			}
			var data []byte
			switch v := args[1].(type) {
			case *value.String:
				data = []byte(v.Value)
			case *value.Bytes:
				data = v.Value
			default:
				return makeError("fwrite: expected string or bytes")
			}
			_, err := h.File.Write(data)
			if err != nil {
				return makeError("fwrite: " + err.Error())
			}
			return value.True
		},
	},
	"fclose": {
		Name: "fclose",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 1 {
				return makeError("fclose: want 1 argument")
			}
			h, ok := args[0].(*value.Handle)
			if !ok {
				return makeError("fclose: expected handle")
			}
			err := h.File.Close()
			if err != nil {
				return makeError("fclose: " + err.Error())
			}
			return value.True
		},
	},
}

func deepEqual(a, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *value.Number:
		bv := b.(*value.Number)
		return av.Value.Cmp(bv.Value) == 0
	case *value.Boolean:
		bv := b.(*value.Boolean)
		return av.Value == bv.Value
	case *value.Rune:
		bv := b.(*value.Rune)
		return av.Value == bv.Value
	case *value.String:
		bv := b.(*value.String)
		return av.Value == bv.Value
	case *value.Symbol:
		bv := b.(*value.Symbol)
		return av.Value == bv.Value
	case *value.List:
		bv := b.(*value.List)
		if len(av.Elements) != len(bv.Elements) {
			return false
		}
		if av.Shape != bv.Shape {
			return false
		}
		for i := range av.Elements {
			if !deepEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func makeOk(v value.Value) value.Value {
	return &value.List{
		Elements: []value.Value{
			&value.Symbol{Value: "ok"},
			v,
		},
		Shape:  "ok",
		Fields: []string{"tag", "value"},
	}
}

func makeError(msg string) value.Value {
	return &value.List{
		Elements: []value.Value{
			&value.Symbol{Value: "error"},
			&value.String{Value: msg},
		},
		Shape:  "error",
		Fields: []string{"tag", "reason"},
	}
}

func hashValue(v value.Value) int64 {
	switch val := v.(type) {
	case *value.Number:
		// Hash the string representation for consistency
		return hashString(val.Inspect())
	case *value.Boolean:
		if val.Value {
			return 1
		}
		return 0
	case *value.Rune:
		return int64(val.Value)
	case *value.String:
		return hashString(val.Value)
	case *value.Symbol:
		return hashString(val.Value)
	case *value.List:
		// Hash shape + elements
		h := hashString(val.Shape)
		for _, e := range val.Elements {
			h = h*31 + hashValue(e)
		}
		return h
	default:
		return 0
	}
}

func hashString(s string) int64 {
	var h int64 = 0
	for _, r := range s {
		h = h*31 + int64(r)
	}
	if h < 0 {
		h = -h
	}
	return h
}
