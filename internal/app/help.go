package app

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/its-haze/league-rpc/internal/logging"
)

// WithLogs wires the in-memory log ring and the logs directory so the Help
// screen can show recent lines, live-tail new ones, and open the folder.
func WithLogs(ring *logging.Ring, logDir string) Option {
	return func(a *App) {
		a.logs = ring
		a.logDir = logDir
	}
}

// GetRecentLogs returns the buffered log lines, oldest first. Empty until
// WithLogs is wired.
func (a *App) GetRecentLogs() []string {
	if a.logs == nil {
		return nil
	}
	return a.logs.RecentLines()
}

// SubscribeLogs returns a channel that receives each new log line as it is
// written. The GUI adapter forwards these to a frontend event.
func (a *App) SubscribeLogs() <-chan string {
	if a.logs == nil {
		return nil
	}
	return a.logs.Subscribe()
}

// OpenLogsFolder opens the logs directory in Explorer.
func (a *App) OpenLogsFolder() error {
	if a.logDir == "" {
		return fmt.Errorf("logs directory unavailable")
	}
	return a.openFolder(a.logDir)
}

// defaultOpenFolder is the real Explorer-launching implementation; tests
// inject a fake so they never shell out.
func defaultOpenFolder(path string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("open folder is not supported on %s", runtime.GOOS)
	}
	return exec.Command("explorer", path).Start()
}

// GetDiagnostics returns a paste-ready Markdown block for bug reports: app
// version, OS, and the current connection state.
func (a *App) GetDiagnostics() string {
	status := a.GetStatus()
	upd := a.GetUpdateStatus()

	lastErr := "none"
	if upd.LastError != "" {
		lastErr = upd.LastError
	}

	return fmt.Sprintf(
		"League RPC diagnostics\n"+
			"- Version: %s\n"+
			"- OS: %s/%s\n"+
			"- League process: %v\n"+
			"- League Client (LCU): %v\n"+
			"- Discord: %v\n"+
			"- Paused: %v\n"+
			"- Game flow phase: %s\n"+
			"- Last update error: %s\n",
		a.GetVersion(), runtime.GOOS, runtime.GOARCH,
		status.LeagueProcess, status.LCUConnected, status.DiscordConnected,
		status.Paused, status.GameFlowPhase, lastErr,
	)
}
