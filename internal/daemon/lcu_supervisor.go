package daemon

import (
	"context"
	"sync/atomic"
	"time"
)

// ProcessChecker reports whether any process in names is currently running.
// internal/process.Checker satisfies this.
type ProcessChecker interface {
	IsRunning(names ...string) (bool, error)
}

// LCUSupervisor wraps a Supervisor for the LCU connection, and separately
// tracks whether League's own process is running. See CONTEXT.md's Connection Supervisor entry.
type LCUSupervisor struct {
	*Supervisor

	checker        ProcessChecker
	processNames   []string
	pollInterval   time.Duration
	leagueDetected atomic.Bool
	connected      atomic.Bool
}

// NewLCUSupervisor builds an LCUSupervisor around connector (typically
// *lcu.Client). processPollInterval controls the League-process check cadence.
func NewLCUSupervisor(
	connector Connector,
	checker ProcessChecker,
	retryInterval, connectPollInterval, processPollInterval time.Duration,
	opts ...Option,
) *LCUSupervisor {
	ls := &LCUSupervisor{
		checker:      checker,
		processNames: leagueProcessNames,
		pollInterval: processPollInterval,
	}
	ls.Supervisor = NewSupervisor(&leagueGatedConnector{Connector: connector, ls: ls}, retryInterval, connectPollInterval, opts...)

	// Wrap the caller's callbacks so Connected() tracks the gated connector,
	// not the raw LCU client's IsConnected(), which never flips back to false.
	userOnConnect := ls.Supervisor.onConnect
	userOnDisconnect := ls.Supervisor.onDisconnect
	ls.Supervisor.onConnect = func() {
		ls.connected.Store(true)
		if userOnConnect != nil {
			userOnConnect()
		}
	}
	ls.Supervisor.onDisconnect = func() {
		ls.connected.Store(false)
		if userOnDisconnect != nil {
			userOnDisconnect()
		}
	}
	return ls
}

// Connected reports whether the LCU is reachable and League's process is
// still running, per the gated connector's last connect/disconnect callback.
func (ls *LCUSupervisor) Connected() bool {
	return ls.connected.Load()
}

// leagueGatedConnector reports connected only while League's process is
// still running, since the underlying Connector has no live disconnect signal of its own.
type leagueGatedConnector struct {
	Connector
	ls *LCUSupervisor
}

func (c *leagueGatedConnector) IsConnected() bool {
	return c.Connector.IsConnected() && c.ls.LeagueProcessDetected()
}

// LeagueProcessDetected reports whether League's own process was running as
// of the last poll.
func (ls *LCUSupervisor) LeagueProcessDetected() bool {
	return ls.leagueDetected.Load()
}

// Run polls for League's process alongside the underlying Supervisor's
// connect-retry loop, and blocks until ctx is canceled.
func (ls *LCUSupervisor) Run(ctx context.Context) {
	// Prime leagueDetected before any connect attempt is evaluated, so the
	// leagueGatedConnector never sees a stale "not detected" default.
	ls.checkLeagueProcess()
	go ls.pollLeagueProcess(ctx)
	ls.Supervisor.Run(ctx)
}

func (ls *LCUSupervisor) pollLeagueProcess(ctx context.Context) {
	ticker := time.NewTicker(ls.pollInterval)
	defer ticker.Stop()

	ls.checkLeagueProcess()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ls.checkLeagueProcess()
		}
	}
}

func (ls *LCUSupervisor) checkLeagueProcess() {
	running, err := ls.checker.IsRunning(ls.processNames...)
	if err != nil {
		// Leave the last known state in place rather than flapping to
		// "not detected" on a transient process-listing error.
		return
	}
	ls.leagueDetected.Store(running)
}
