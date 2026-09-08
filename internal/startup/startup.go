// Package startup keeps the Windows "start at logon" registry entry in step
// with the Behavior.LaunchAtStartup setting.
package startup

import (
	"fmt"
	"os"
	"slices"
)

// ValueName is the HKCU\...\Run entry this app owns.
const ValueName = "LeagueRPC"

// HiddenArg is appended to the Run command. A run started by the Run entry
// carries it and opens straight to the tray; a manual run does not.
const HiddenArg = "--hidden"

// RunKey is the slice of the Windows Run registry key the reconciler touches.
// The Windows build backs it with the real key; tests pass a fake.
type RunKey interface {
	// Value returns the current string value and whether it is set.
	Value(name string) (string, bool, error)
	SetValue(name, value string) error
	DeleteValue(name string) error
}

// Reconciler drives one Run value toward a desired on/off state.
type Reconciler struct {
	key RunKey
	// exePath is a field so tests don't depend on os.Executable.
	exePath func() (string, error)
}

// New builds a Reconciler over key.
func New(key RunKey) *Reconciler {
	return &Reconciler{key: key, exePath: os.Executable}
}

// Enabled reports whether the Run value is currently set.
func (r *Reconciler) Enabled() (bool, error) {
	_, ok, err := r.key.Value(ValueName)
	return ok, err
}

// Reconcile makes the Run value match want: this executable plus the hidden
// marker when true, deleted when false. Idempotent; writes only on a diff.
func (r *Reconciler) Reconcile(want bool) error {
	cur, ok, err := r.key.Value(ValueName)
	if err != nil {
		return err
	}

	if !want {
		if !ok {
			return nil
		}
		return r.key.DeleteValue(ValueName)
	}

	exe, err := r.exePath()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	desired := Command(exe)
	if ok && cur == desired {
		return nil
	}
	return r.key.SetValue(ValueName, desired)
}

// Command is the string stored in the Run value: the quoted executable path
// followed by the hidden marker.
func Command(exe string) string {
	return `"` + exe + `" ` + HiddenArg
}

// StartedHidden reports whether args (pass os.Args[1:]) carries the hidden
// marker, meaning the Run entry launched this process.
func StartedHidden(args []string) bool {
	return slices.Contains(args, HiddenArg)
}
