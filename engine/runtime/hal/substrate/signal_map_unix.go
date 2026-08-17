//go:build !windows

package substrate

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

var errUnsupportedSignal = errors.New("unsupported portable signal")

func hostSignalForWatch(name string) (os.Signal, bool) {
	switch name {
	case "interrupt":
		return os.Interrupt, true
	case "terminate":
		return syscall.SIGTERM, true
	default:
		return nil, false
	}
}

func portableSignalName(sig os.Signal) (string, bool) {
	switch sig {
	case os.Interrupt:
		return "interrupt", true
	case syscall.SIGTERM:
		return "terminate", true
	default:
		return "", false
	}
}

func sendPortableSignal(process *os.Process, name string) error {
	sig, ok := hostSignalForWatch(name)
	if !ok {
		return fmt.Errorf("%w: %s", errUnsupportedSignal, name)
	}
	return process.Signal(sig)
}

func isUnsupportedSignal(err error) bool { return errors.Is(err, errUnsupportedSignal) }
