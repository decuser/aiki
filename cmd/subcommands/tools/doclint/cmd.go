package doclint

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	fmt.Fprintln(os.Stderr, "doclint: not yet ported to new engine")
	return 1
}
