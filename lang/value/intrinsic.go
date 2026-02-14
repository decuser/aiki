package value

import "fmt"

// Intrinsic represents an evaluator-level primitive that needs
// access to the environment, AST node, or other evaluation context.
// Unlike Builtins (pure Go functions), Intrinsics are language mechanics.
//
// Examples: load, apply, import, export
type Intrinsic struct {
	Name string
}

func (i *Intrinsic) Type() Type      { return IntrinsicType }
func (i *Intrinsic) Inspect() string { return fmt.Sprintf("[@intrinsic, %s]", i.Name) }

// Intrinsics is the fixed set of evaluator-level primitives.
var Intrinsics = map[string]*Intrinsic{
	"load":   {Name: "load"},
	"apply":  {Name: "apply"},
	"import": {Name: "import"},
	"export": {Name: "export"},
}
