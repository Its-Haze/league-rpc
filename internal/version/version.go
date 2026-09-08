// Package version reports the running build's version. The real value is
// injected at build time; an un-injected build is treated as a dev build.
package version

import "strings"

// injected is set via -ldflags -X on this var's fully-qualified name; empty
// for `go build`/`go run` and every local dev build.
var injected = ""

// DevPlaceholder is what Version reports when no release tag was injected.
const DevPlaceholder = "dev (unreleased build)"

// Version returns the running build's version: the injected release tag with
// any leading "v" trimmed, or DevPlaceholder for an un-injected build.
func Version() string { return resolve(injected) }

// IsDev reports whether this build carries no injected release tag; the App
// Update flow uses it to disable checks entirely for a local build.
func IsDev() bool { return strings.TrimSpace(injected) == "" }

func resolve(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DevPlaceholder
	}
	return strings.TrimPrefix(raw, "v")
}
