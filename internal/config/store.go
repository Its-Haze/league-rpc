package config

import (
	"sync"
	"sync/atomic"
)

// Store holds the live configuration. Readers call Load for a lock-free
type Store struct {
	cur  atomic.Pointer[Config]
	mu   sync.Mutex // serializes writers and the subs slice
	subs []chan *Config
}

// NewStore wraps initial. It does not validate: Load must always return
// something, and the load path already clamped whatever came off disk.
func NewStore(initial *Config) *Store {
	s := &Store{}
	s.cur.Store(initial)
	return s
}

// Load returns the current config snapshot. Never nil. Do not mutate it;
// Apply always allocates a fresh Config, so the pointer you hold is stable.
func (s *Store) Load() *Config {
	return s.cur.Load()
}

// Subscribe returns a channel that receives the new snapshot on every Apply.
func (s *Store) Subscribe() <-chan *Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan *Config, 1)
	s.subs = append(s.subs, ch)
	return ch
}

// Apply validates next, writes it to disk, then swaps it in and notifies
// subscribers. On a validation or write error nothing changes.
func (s *Store) Apply(next Config) error {
	if err := next.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := Save(&next); err != nil {
		return err
	}

	n := &next
	s.cur.Store(n)
	for _, ch := range s.subs {
		// Drop any stale pending value, then push the latest. Both sends
		// are non-blocking so Apply returns regardless of subscriber speed.
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- n:
		default:
		}
	}
	return nil
}
