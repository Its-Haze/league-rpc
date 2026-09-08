// Package daemon implements the Daemon's core lifecycle.
// See ADR-0002.
package daemon

import (
	"context"
	"errors"
	"time"
)

var errConnectFailed = errors.New("connect failed")

// Connector is anything a Supervisor can retry-connect to and monitor.
// discord.Client and lcu.Client both already satisfy this.
type Connector interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
}

// Gate optionally blocks connection attempts until it reports ready.
// The Discord supervisor uses this to wait for Discord's process.
type Gate interface {
	Ready() (bool, error)
}

// Option configures a Supervisor.
type Option func(*Supervisor)

// WithGate sets a Gate that must report ready before each Connect() attempt.
func WithGate(g Gate) Option {
	return func(s *Supervisor) { s.gate = g }
}

// WithOnConnect sets a callback invoked every time Connect() succeeds.
func WithOnConnect(f func()) Option {
	return func(s *Supervisor) { s.onConnect = f }
}

// WithOnDisconnect sets a callback invoked every time the connection is torn
// down, whether by a detected disconnect or by shutdown.
func WithOnDisconnect(f func()) Option {
	return func(s *Supervisor) { s.onDisconnect = f }
}

// Supervisor retries Connect() on a Connector forever and never returns
// except when ctx is canceled. See ADR-0002.
type Supervisor struct {
	connector     Connector
	gate          Gate // nil means always ready
	retryInterval time.Duration
	pollInterval  time.Duration // how often IsConnected() is polled once connected

	onConnect    func()
	onDisconnect func()
}

// NewSupervisor builds a Supervisor around connector. retryInterval controls
// retry cadence; pollInterval controls how often IsConnected() is checked.
func NewSupervisor(connector Connector, retryInterval, pollInterval time.Duration, opts ...Option) *Supervisor {
	s := &Supervisor{
		connector:     connector,
		retryInterval: retryInterval,
		pollInterval:  pollInterval,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run retries forever until ctx is canceled. It never returns for any
// other reason.
func (s *Supervisor) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		if s.gate != nil {
			ready, err := s.gate.Ready()
			if err != nil || !ready {
				if !sleep(ctx, s.retryInterval) {
					return
				}
				continue
			}
		}

		connectErr, canceled := s.connect(ctx)
		if canceled {
			// Connect() was still in flight and can't be interrupted; abandon it.
			return
		}
		if connectErr != nil {
			if !sleep(ctx, s.retryInterval) {
				return
			}
			continue
		}

		if s.onConnect != nil {
			s.onConnect()
		}

		canceled = s.waitWhileConnected(ctx)

		_ = s.connector.Disconnect()
		if s.onDisconnect != nil {
			s.onDisconnect()
		}

		if canceled {
			return
		}
		// Otherwise a disconnect was detected: loop back and retry connecting.
	}
}

// Connected reports whether the connector currently reports itself connected.
func (s *Supervisor) Connected() bool {
	return s.connector.IsConnected()
}

// connect runs Connect() in a goroutine so ctx cancellation can be observed
// even while Connect() is still blocked. See ADR-0002.
func (s *Supervisor) connect(ctx context.Context) (connectErr error, canceled bool) {
	done := make(chan error, 1)
	go func() {
		done <- s.connector.Connect()
	}()

	select {
	case <-ctx.Done():
		return nil, true
	case connectErr = <-done:
		return connectErr, false
	}
}

// waitWhileConnected polls IsConnected() until it goes false or ctx is
// canceled. Returns true if ctx was canceled, false on a detected disconnect.
func (s *Supervisor) waitWhileConnected(ctx context.Context) bool {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return true
		case <-ticker.C:
			if !s.connector.IsConnected() {
				return false
			}
		}
	}
}

// sleep waits for d or until ctx is canceled. Returns false if ctx was
// canceled first.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
