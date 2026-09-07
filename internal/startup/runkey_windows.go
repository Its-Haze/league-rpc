//go:build windows

package startup

import (
	"errors"

	"golang.org/x/sys/windows/registry"
)

// runKeyPath is the per-user autorun key. Writing here needs no elevation.
const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// SystemRunKey returns the real HKCU Run key.
func SystemRunKey() RunKey { return windowsRunKey{} }

type windowsRunKey struct{}

func (windowsRunKey) Value(name string) (string, bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", false, err
	}
	defer k.Close()

	v, _, err := k.GetStringValue(name)
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (windowsRunKey) SetValue(name, value string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(name, value)
}

func (windowsRunKey) DeleteValue(name string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.DeleteValue(name); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}
