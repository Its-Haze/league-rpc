//go:build !windows

package startup

import "errors"

// errUnsupported is returned only when a non-Windows build tries to actually
// write the Run key; reads report "not set" so the disable path is a no-op.
var errUnsupported = errors.New("startup: Run key is only supported on Windows")

// SystemRunKey returns a stub used on non-Windows builds. Reads succeed as
// "not set"; writes fail, so enabling start-with-Windows there is a loud error.
func SystemRunKey() RunKey { return unsupportedRunKey{} }

type unsupportedRunKey struct{}

func (unsupportedRunKey) Value(string) (string, bool, error) { return "", false, nil }
func (unsupportedRunKey) SetValue(string, string) error      { return errUnsupported }
func (unsupportedRunKey) DeleteValue(string) error           { return errUnsupported }
