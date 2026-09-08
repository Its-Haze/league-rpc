package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeProcessChecker struct {
	running bool
	err     error
}

func (f *fakeProcessChecker) IsRunning(names ...string) (bool, error) {
	return f.running, f.err
}

func TestLCUSupervisor_LeagueProcessDetected_TracksChecker(t *testing.T) {
	connector := &fakeConnector{connectFailures: 1 << 30} // LCU never connects in this test
	checker := &fakeProcessChecker{running: false}

	s := NewLCUSupervisor(connector, checker, testRetryInterval, testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go s.Run(ctx)

	if s.LeagueProcessDetected() {
		t.Fatal("expected LeagueProcessDetected() to start false")
	}

	checker.running = true
	waitFor(t, testTimeout, s.LeagueProcessDetected)

	checker.running = false
	waitFor(t, testTimeout, func() bool { return !s.LeagueProcessDetected() })
}

func TestLCUSupervisor_ProcessDetectionIndependentOfConnectSuccess(t *testing.T) {
	// Process detection must not depend on the LCU having connected yet.
	connector := &fakeConnector{connectFailures: 1 << 30}
	checker := &fakeProcessChecker{running: true}

	s := NewLCUSupervisor(connector, checker, testRetryInterval, testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go s.Run(ctx)

	waitFor(t, testTimeout, s.LeagueProcessDetected)

	if connector.connected.Load() {
		t.Fatal("connector should not report connected in this test")
	}
}

func TestLCUSupervisor_ChecksErrorLeavesLastKnownState(t *testing.T) {
	connector := &fakeConnector{connectFailures: 1 << 30}
	checker := &fakeProcessChecker{running: true}

	s := NewLCUSupervisor(connector, checker, testRetryInterval, testPollInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go s.Run(ctx)

	waitFor(t, testTimeout, s.LeagueProcessDetected)

	checker.err = errors.New("boom")
	time.Sleep(5 * testPollInterval)

	if !s.LeagueProcessDetected() {
		t.Fatal("expected last known state to be preserved when the checker errors")
	}
}

func TestLCUSupervisor_DetectsDisconnectWhenLeagueProcessDisappears(t *testing.T) {
	// Mirrors the real lcu.Client: IsConnected() stays true forever once
	// Connect() has succeeded, since lcu-gopher has no live disconnect
	// signal. LCUSupervisor must still detect the disconnect by noticing
	// League's process is gone.
	connector := &fakeConnector{}
	checker := &fakeProcessChecker{running: true}

	var connectCount atomic.Int32
	s := NewLCUSupervisor(connector, checker, testRetryInterval, testPollInterval, testPollInterval,
		WithOnConnect(func() { connectCount.Add(1) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go s.Run(ctx)

	waitFor(t, testTimeout, func() bool { return connectCount.Load() == 1 })

	if !connector.IsConnected() {
		t.Fatal("underlying connector should still report connected, as the real lcu.Client would")
	}

	checker.running = false // League's process disappears; connector.IsConnected() never flips

	waitFor(t, testTimeout, func() bool { return connectCount.Load() == 2 })
}

func TestLCUSupervisor_StillConnectsAndDisconnectsLikeSupervisor(t *testing.T) {
	connector := &fakeConnector{}
	checker := &fakeProcessChecker{running: true}
	var connected bool
	s := NewLCUSupervisor(connector, checker, testRetryInterval, testPollInterval, testPollInterval,
		WithOnConnect(func() { connected = true }))

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	waitFor(t, testTimeout, func() bool { return connected })

	cancel()
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Run did not return after ctx cancellation")
	}
}
