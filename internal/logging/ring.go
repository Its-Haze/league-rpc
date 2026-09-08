package logging

import (
	"strings"
	"sync"
)

// DefaultRingSize is how many recent log lines the in-memory buffer keeps.
const DefaultRingSize = 500

// Ring is a fixed-capacity, concurrency-safe buffer of the most recent log
type Ring struct {
	mu   sync.Mutex
	buf  []string
	next int
	full bool
	subs []chan string
}

// NewRing builds a Ring holding up to size lines (size <= 0 uses the default).
func NewRing(size int) *Ring {
	if size <= 0 {
		size = DefaultRingSize
	}
	return &Ring{buf: make([]string, size)}
}

// Write records p as one line and fans it out to subscribers. It never errors
// and never blocks on a slow subscriber.
func (r *Ring) Write(p []byte) (int, error) {
	line := strings.TrimRight(string(p), "\n")

	r.mu.Lock()
	r.buf[r.next] = line
	r.next = (r.next + 1) % len(r.buf)
	if r.next == 0 {
		r.full = true
	}
	subs := append([]chan string(nil), r.subs...)
	r.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- line:
		default:
		}
	}
	return len(p), nil
}

// RecentLines returns the buffered lines oldest-first.
func (r *Ring) RecentLines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.full {
		out := make([]string, r.next)
		copy(out, r.buf[:r.next])
		return out
	}
	out := make([]string, 0, len(r.buf))
	out = append(out, r.buf[r.next:]...)
	out = append(out, r.buf[:r.next]...)
	return out
}

// Subscribe returns a channel that receives each new line as it is written.
func (r *Ring) Subscribe() <-chan string {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch := make(chan string, 64)
	r.subs = append(r.subs, ch)
	return ch
}
