package lint

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	fmt.Fprintln(os.Stderr, "lint: not yet ported to new engine")
	return 1
}
