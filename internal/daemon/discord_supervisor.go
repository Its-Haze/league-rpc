package daemon

import (
	"context"
	"sync/atomic"
	"time"
)

// LeagueDetector reports whether League's own process is running.
// *LCUSupervisor satisfies this.
type LeagueDetector interface {
	LeagueProcessDetected() bool
}

// DiscordSupervisor wraps a Supervisor for the Discord IPC connection, also
// gated on League's own process via league. See ADR-0003.
type DiscordSupervisor struct {
	*Supervisor

	checker      ProcessChecker
	processNames []string
	pollInterval time.Duration
	league       LeagueDetector
	processUp    atomic.Bool
	connected    atomic.Bool
}

// newDiscordSupervisor builds a DiscordSupervisor; league gates connect/stay-connected.
func newDiscordSupervisor(
	connector Connector,
	checker ProcessChecker,
	league LeagueDetector,
	retryInterval, connectPollInterval, processPollInterval time.Duration,
	opts ...Option,
) *DiscordSupervisor {
	ds := &DiscordSupervisor{
		checker:      checker,
		processNames: discordProcessNames,
		pollInterval: processPollInterval,
		league:       league,
	}
	ds.Supervisor = NewSupervisor(&discordGatedConnector{Connector: connector, ds: ds}, retryInterval, connectPollInterval, opts...)

	// Wrap the caller's callbacks so Connected() tracks the gated connector,
	// not the raw discord.Client's IsConnected(), which never flips back to false.
	userOnConnect := ds.Supervisor.onConnect
	userOnDisconnect := ds.Supervisor.onDisconnect
	ds.Supervisor.onConnect = func() {
		ds.connected.Store(true)
		if userOnConnect != nil {
			userOnConnect()
		}
	}
	ds.Supervisor.onDisconnect = func() {
		ds.connected.Store(false)
		if userOnDisconnect != nil {
			userOnDisconnect()
		}
	}
	return ds
}

// Connected reports whether Discord IPC is reachable and Discord's process is
// still running, per the gated connector's last connect/disconnect callback.
func (ds *DiscordSupervisor) Connected() bool {
	return ds.connected.Load()
}

// discordGatedConnector reports connected only while Discord's process is
// still running, since the underlying Connector has no live disconnect signal of its own.
type discordGatedConnector struct {
	Connector
	ds *DiscordSupervisor
}

func (c *discordGatedConnector) IsConnected() bool {
	return c.Connector.IsConnected() && c.ds.processRunning() && c.ds.league.LeagueProcessDetected()
}

func (ds *DiscordSupervisor) processRunning() bool {
	return ds.processUp.Load()
}

// Run polls for Discord's process alongside the underlying Supervisor's
// connect-retry loop, and blocks until ctx is canceled.
func (ds *DiscordSupervisor) Run(ctx context.Context) {
	// Prime processUp before any connect attempt is evaluated, so the
	// discordGatedConnector never sees a stale "not detected" default.
	ds.checkProcess()
	go ds.pollProcess(ctx)
	ds.Supervisor.Run(ctx)
}

func (ds *DiscordSupervisor) pollProcess(ctx context.Context) {
	ticker := time.NewTicker(ds.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ds.checkProcess()
		}
	}
}

func (ds *DiscordSupervisor) checkProcess() {
	running, err := ds.checker.IsRunning(ds.processNames...)
	if err != nil {
		// Leave the last known state in place rather than flapping to
		// "not detected" on a transient process-listing error.
		return
	}
	ds.processUp.Store(running)
}
