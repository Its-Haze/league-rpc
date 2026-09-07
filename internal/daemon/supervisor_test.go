package daemon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

const (
	testRetryInterval = 2 * time.Millisecond
	testPollInterval  = 2 * time.Millisecond
	testTimeout       = 2 * time.Second
)

type fakeConnector struct {
	connectFailures int32 // number of times Connect() should fail before succeeding
	connectCalls    atomic.Int32
	connected       atomic.Bool
	disconnectCalls atomic.Int32
}

func (f *fakeConnector) Connect() error {
	n := f.connectCalls.Add(1)
	if n <= f.connectFailures {
		return errConnectFailed
	}
	f.connected.Store(true)
	return nil
}

func (f *fakeConnector) Disconnect() error {
	f.disconnectCalls.Add(1)
	f.connected.Store(false)
	return nil
}

func (f *fakeConnector) IsConnected() bool {
	return f.connected.Load()
}

// blockingConnector simulates lcu.Client.Connect(), which has no
// cancellation hook and blocks until League actually launches.
type blockingConnector struct {
	unblock chan struct{}
}

func (f *blockingConnector) Connect() error {
	<-f.unblock
	return nil
}

func (f *blockingConnector) Disconnect() error { return nil }
func (f *blockingConnector) IsConnected() bool { return false }

type fakeGate struct {
	ready atomic.Bool
}

func (g *fakeGate) Ready() (bool, error) {
	return g.ready.Load(), nil
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

func TestSupervisor_ConnectsOnFirstSuccess(t *testing.T) {
	connector := &fakeConnector{}
	var connected atomic.Bool
	s := NewSupervisor(connector, testRetryInterval, testPollInterval,
		WithOnConnect(func() { connected.Store(true) }))

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	waitFor(t, testTimeout, connected.Load)

	cancel()
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Run did not return after ctx cancellation")
	}

	if connector.disconnectCalls.Load() == 0 {
		t.Error("expected Disconnect to be called on shutdown")
	}
}

func TestSupervisor_RetriesUntilConnectSucceeds(t *testing.T) {
	connector := &fakeConnector{connectFailures: 3}
	var connected atomic.Bool
	s := NewSupervisor(connector, testRetryInterval, testPollInterval,
		WithOnConnect(func() { connected.Store(true) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	waitFor(t, testTimeout, connected.Load)

	if got := connector.connectCalls.Load(); got < 4 {
		t.Errorf("expected at least 4 Connect() attempts, got %d", got)
	}

	cancel()
	<-done
}

func TestSupervisor_NeverExitsOnRepeatedFailure(t *testing.T) {
	connector := &fakeConnector{connectFailures: 1 << 30} // never succeeds
	s := NewSupervisor(connector, testRetryInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	// Let several retry cycles happen; Run must still be blocked.
	time.Sleep(20 * testRetryInterval)
	select {
	case <-done:
		t.Fatal("Run returned despite repeated connection failure")
	default:
	}

	cancel()
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Run did not return after ctx cancellation")
	}
}

func TestSupervisor_WaitsForGateBeforeConnecting(t *testing.T) {
	connector := &fakeConnector{}
	gate := &fakeGate{}
	var connected atomic.Bool
	s := NewSupervisor(connector, testRetryInterval, testPollInterval,
		WithGate(gate),
		WithOnConnect(func() { connected.Store(true) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go s.Run(ctx)

	// Gate isn't ready yet: Connect must not be attempted.
	time.Sleep(10 * testRetryInterval)
	if connector.connectCalls.Load() != 0 {
		t.Fatalf("expected no Connect() attempts while gate isn't ready, got %d", connector.connectCalls.Load())
	}

	gate.ready.Store(true)
	waitFor(t, testTimeout, connected.Load)
}

func TestSupervisor_ReconnectsAfterDisconnectIsDetected(t *testing.T) {
	connector := &fakeConnector{}
	var connectCount atomic.Int32
	s := NewSupervisor(connector, testRetryInterval, testPollInterval,
		WithOnConnect(func() { connectCount.Add(1) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go s.Run(ctx)

	waitFor(t, testTimeout, func() bool { return connectCount.Load() == 1 })

	// Simulate the connection dropping out from under the supervisor.
	connector.connected.Store(false)

	waitFor(t, testTimeout, func() bool { return connectCount.Load() == 2 })
}

func TestSupervisor_RunReturnsPromptlyEvenIfConnectNeverReturns(t *testing.T) {
	connector := &blockingConnector{unblock: make(chan struct{})} // never unblocked
	s := NewSupervisor(connector, testRetryInterval, testPollInterval)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	// Let Run actually enter the blocked Connect() call before canceling.
	time.Sleep(10 * testRetryInterval)
	cancel()

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("Run did not return after ctx cancellation, even though Connect() was still blocked")
	}
}

func TestSupervisor_DisconnectDoesNotBlockOtherSupervisor(t *testing.T) {
	stuck := &fakeConnector{connectFailures: 1 << 30}
	healthy := &fakeConnector{}

	var healthyConnected atomic.Bool
	stuckSupervisor := NewSupervisor(stuck, testRetryInterval, testPollInterval)
	healthySupervisor := NewSupervisor(healthy, testRetryInterval, testPollInterval,
		WithOnConnect(func() { healthyConnected.Store(true) }))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go stuckSupervisor.Run(ctx)
	go healthySupervisor.Run(ctx)

	waitFor(t, testTimeout, healthyConnected.Load)
}
