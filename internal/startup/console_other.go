//go:build !windows

package startup

// DetachConsole is a no-op away from Windows.
func DetachConsole() {}
