package logging

import (
	"fmt"
	"sync"
	"testing"
)

func TestRing_WrapsAtCapacity(t *testing.T) {
	r := NewRing(3)
	for i := 0; i < 5; i++ {
		fmt.Fprintf(r, "line %d\n", i)
	}

	got := r.RecentLines()
	want := []string{"line 2", "line 3", "line 4"}
	if len(got) != len(want) {
		t.Fatalf("RecentLines len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRing_RecentLinesBeforeFull(t *testing.T) {
	r := NewRing(10)
	fmt.Fprintln(r, "only")

	got := r.RecentLines()
	if len(got) != 1 || got[0] != "only" {
		t.Fatalf("RecentLines = %v, want [only]", got)
	}
}

func TestRing_ConcurrentAppendAndReadIsRaceFree(t *testing.T) {
	r := NewRing(64)
	var wg sync.WaitGroup

	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				fmt.Fprintf(r, "w%d-%d\n", w, i)
			}
		}(w)
	}
	for reader := 0; reader < 4; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = r.RecentLines()
			}
		}()
	}
	wg.Wait()

	if n := len(r.RecentLines()); n != 64 {
		t.Fatalf("buffer should be full at 64, got %d", n)
	}
}

func TestRing_SubscribeReceivesNewLines(t *testing.T) {
	r := NewRing(10)
	sub := r.Subscribe()

	fmt.Fprintln(r, "hello")

	select {
	case got := <-sub:
		if got != "hello" {
			t.Fatalf("subscriber got %q, want hello", got)
		}
	default:
		t.Fatal("subscriber received nothing")
	}
}
