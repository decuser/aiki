package value

import "fmt"

// FileLock is an opaque runtime-owned exclusive interprocess lock token.
type FileLock struct{ ID uint64 }

func (l *FileLock) Type() Type { return FileLockType }
func (l *FileLock) Inspect() string {
	if l == nil || l.ID == 0 {
		return "<file-lock>"
	}
	return fmt.Sprintf("<file-lock:%d>", l.ID)
}
