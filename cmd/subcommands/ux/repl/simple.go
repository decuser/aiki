package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type SimpleReader struct {
	scanner *bufio.Scanner
}

func NewSimpleReader() *SimpleReader {
	return &SimpleReader{scanner: bufio.NewScanner(os.Stdin)}
}

func (r *SimpleReader) Prompt(prompt string) (string, error) {
	fmt.Print(prompt)
	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}
	if err := r.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (r *SimpleReader) Close() {}
