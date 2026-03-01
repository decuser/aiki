package debug

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	fmt.Fprintln(os.Stderr, "debug: not yet ported to new engine")
	return 1
}
