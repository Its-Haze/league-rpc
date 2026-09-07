//go:build !windows

package main

// workAreaDIP has no implementation away from Windows, so the preferred size
// is used as-is.
func workAreaDIP() (int, int, bool) { return 0, 0, false }
