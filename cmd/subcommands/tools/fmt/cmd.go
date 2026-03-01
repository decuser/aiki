package fmt

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	fmt.Fprintln(os.Stderr, "fmt: not yet ported to new engine")
	return 1
}
