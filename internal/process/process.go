// Package process detects whether a named OS process (Discord, League) is
// currently running.
package process

import (
	"strings"

	gopsutilprocess "github.com/shirou/gopsutil/v4/process"
)

// Lister lists the names of currently running processes.
// Faked in tests, so no test depends on a real process.
type Lister interface {
	ProcessNames() ([]string, error)
}

type gopsutilLister struct{}

func (gopsutilLister) ProcessNames() ([]string, error) {
	procs, err := gopsutilprocess.Processes()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(procs))
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			// Processes can exit between listing and inspection; skip them
			// rather than failing the whole check.
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// Checker checks whether any of a set of process names is currently running.
type Checker struct {
	lister Lister
}

// NewChecker returns a Checker backed by the real OS process list.
func NewChecker() *Checker {
	return &Checker{lister: gopsutilLister{}}
}

// NewCheckerWithLister returns a Checker backed by lister, for injecting a
// fake in tests.
func NewCheckerWithLister(lister Lister) *Checker {
	return &Checker{lister: lister}
}

// IsRunning reports whether any process in names is currently running.
// Matching is case-insensitive.
func (c *Checker) IsRunning(names ...string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}

	running, err := c.lister.ProcessNames()
	if err != nil {
		return false, err
	}

	want := make(map[string]struct{}, len(names))
	for _, n := range names {
		want[strings.ToLower(n)] = struct{}{}
	}

	for _, n := range running {
		if _, ok := want[strings.ToLower(n)]; ok {
			return true, nil
		}
	}
	return false, nil
}
