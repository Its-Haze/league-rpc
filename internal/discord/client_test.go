package discord

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/its-haze/league-rpc/internal/config"
	"github.com/rs/zerolog"
)

// fakeIpcConn stands in for a real named-pipe connection, letting tests
// simulate a send failing (e.g. Discord's Ctrl+R reload) without a real pipe.
type fakeIpcConn struct {
	sendErr   error
	closed    bool
	sendCalls int
	sent      [][]byte
}

func (f *fakeIpcConn) Send(opcode int32, payload []byte) ([]byte, error) {
	f.sendCalls++
	f.sent = append(f.sent, append([]byte(nil), payload...))
	if f.sendErr != nil {
		return nil, f.sendErr
	}
	return []byte(`{"cmd":"SET_ACTIVITY","data":{"code":1000}}`), nil
}

func (f *fakeIpcConn) Close() error {
	f.closed = true
	return nil
}

func newTestClient(t *testing.T, conns ...*fakeIpcConn) *Client {
	t.Helper()
	i := 0
	c := NewClient(config.NewStore(config.DefaultConfig()), zerolog.Nop())
	c.dial = func() (ipcConn, error) {
		if i >= len(conns) {
			t.Fatalf("dial called more times than test provided fake connections")
		}
		conn := conns[i]
		i++
		return conn, nil
	}
	return c
}

// Discord only drops an activity on a null; an empty activity object still
// registers one, showing the bare app name with a timer.
func TestClient_ClearPresenceSendsNullActivity(t *testing.T) {
	conn := &fakeIpcConn{}
	c := newTestClient(t, conn)

	if err := c.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	if err := c.UpdatePresence(&RPCData{Details: "In Game"}); err != nil {
		t.Fatalf("UpdatePresence() failed: %v", err)
	}
	if err := c.ClearPresence(); err != nil {
		t.Fatalf("ClearPresence() failed: %v", err)
	}

	var frame struct {
		Args struct {
			Activity *json.RawMessage `json:"activity"`
		} `json:"args"`
	}
	last := conn.sent[len(conn.sent)-1]
	if err := json.Unmarshal(last, &frame); err != nil {
		t.Fatalf("failed to parse the SET_ACTIVITY frame: %v", err)
	}
	if frame.Args.Activity != nil && string(*frame.Args.Activity) != "null" {
		t.Fatalf("expected a null activity on clear, got %s", *frame.Args.Activity)
	}
}

func TestClient_ConnectThenUpdatePresenceSucceeds(t *testing.T) {
	conn := &fakeIpcConn{}
	c := newTestClient(t, conn)

	if err := c.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	if !c.IsConnected() {
		t.Fatal("expected IsConnected() true after Connect()")
	}

	if err := c.UpdatePresence(&RPCData{Details: "In Game", State: "0/0/0"}); err != nil {
		t.Fatalf("UpdatePresence() failed: %v", err)
	}
	if conn.sendCalls != 2 { // handshake + activity
		t.Fatalf("expected 2 sends (handshake + activity), got %d", conn.sendCalls)
	}
}

func TestClient_SendFailureMarksDisconnectedImmediately(t *testing.T) {
	// Regression: Discord's Ctrl+R reload closes the pipe without killing
	// Discord.exe, so nothing but a failed send itself can signal this.
	conn := &fakeIpcConn{}
	c := newTestClient(t, conn)

	if err := c.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	conn.sendErr = errors.New("The pipe is being closed.")
	if err := c.UpdatePresence(&RPCData{Details: "In Game"}); err == nil {
		t.Fatal("expected UpdatePresence() to fail once the pipe write fails")
	}

	if c.IsConnected() {
		t.Fatal("expected IsConnected() false immediately after a failed send, not after a process check")
	}

	// A second call must fail fast (not connected) rather than reuse the dead conn.
	if err := c.UpdatePresence(&RPCData{Details: "In Game"}); err == nil {
		t.Fatal("expected UpdatePresence() to keep failing while disconnected")
	}
}

func TestClient_ReconnectsWithFreshConnectionAfterDisconnect(t *testing.T) {
	first := &fakeIpcConn{}
	second := &fakeIpcConn{}
	c := newTestClient(t, first, second)

	if err := c.Connect(); err != nil {
		t.Fatalf("first Connect() failed: %v", err)
	}
	if err := c.Disconnect(); err != nil {
		t.Fatalf("Disconnect() failed: %v", err)
	}
	if !first.closed {
		t.Fatal("expected the first connection to be closed on Disconnect()")
	}
	if c.IsConnected() {
		t.Fatal("expected IsConnected() false after Disconnect()")
	}

	if err := c.Connect(); err != nil {
		t.Fatalf("second Connect() failed: %v", err)
	}
	if !c.IsConnected() {
		t.Fatal("expected IsConnected() true after reconnecting")
	}
	if err := c.UpdatePresence(&RPCData{Details: "In Game"}); err != nil {
		t.Fatalf("UpdatePresence() on the new connection failed: %v", err)
	}
	if second.sendCalls == 0 {
		t.Fatal("expected sends to go through the new connection")
	}
}

func TestClient_ClearPresenceNoopsWhenNotConnected(t *testing.T) {
	c := newTestClient(t)
	if err := c.ClearPresence(); err != nil {
		t.Fatalf("expected ClearPresence() to no-op without error when disconnected, got %v", err)
	}
}
