//go:build windows

package substrate

import (
	"errors"
	"fmt"
	"os"
)

var errUnsupportedSignal = errors.New("unsupported portable signal")

func hostSignalForWatch(name string) (os.Signal, bool) {
	if name == "interrupt" {
		return os.Interrupt, true
	}
	return nil, false
}

func portableSignalName(sig os.Signal) (string, bool) {
	if sig == os.Interrupt {
		return "interrupt", true
	}
	return "", false
}

func sendPortableSignal(process *os.Process, name string) error {
	switch name {
	case "terminate":
		return process.Kill()
	default:
		return fmt.Errorf("%w: %s", errUnsupportedSignal, name)
	}
}

func isUnsupportedSignal(err error) bool { return errors.Is(err, errUnsupportedSignal) }
