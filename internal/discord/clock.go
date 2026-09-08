package discord

import "time"

// Ticker abstracts a repeating timer so Updater's heartbeat/reclaim cadence
// can be driven by a fake in tests instead of real wall-clock waits.
type Ticker interface {
	C() <-chan time.Time
	Reset(d time.Duration)
	Stop()
}

// Clock creates Tickers. *realClock wraps time.Ticker for production use.
type Clock interface {
	NewTicker(d time.Duration) Ticker
}

type realClock struct{}

// NewRealClock returns a Clock backed by the real time package.
func NewRealClock() Clock {
	return realClock{}
}

func (realClock) NewTicker(d time.Duration) Ticker {
	return &realTicker{t: time.NewTicker(d)}
}

type realTicker struct {
	t *time.Ticker
}

func (r *realTicker) C() <-chan time.Time   { return r.t.C }
func (r *realTicker) Reset(d time.Duration) { r.t.Reset(d) }
func (r *realTicker) Stop()                 { r.t.Stop() }
