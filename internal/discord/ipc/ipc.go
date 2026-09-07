// Package ipc is a minimal Discord IPC transport: dial the well-known named
// pipe, send a length-prefixed frame, get the response back or a real error.
package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	npipe "gopkg.in/natefinch/npipe.v2"
)

const (
	pipeName    = `\\.\pipe\discord-ipc-0`
	dialTimeout = 2 * time.Second
	readBufSize = 4096
)

// Conn is a single Discord IPC connection. Each Conn owns its own pipe, so
// closing and dialing again needs no shared/global state reset.
type Conn struct {
	pipe *npipe.PipeConn
}

// Dial opens the Discord IPC named pipe.
func Dial() (*Conn, error) {
	p, err := npipe.DialTimeout(pipeName, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("dial Discord IPC pipe: %w", err)
	}
	return &Conn{pipe: p}, nil
}

// Close closes the underlying pipe.
func (c *Conn) Close() error {
	return c.pipe.Close()
}

// Send writes one opcode+payload frame and returns the response payload,
// with the 8-byte opcode/length response header stripped off.
func (c *Conn) Send(opcode int32, payload []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, opcode)
	_ = binary.Write(buf, binary.LittleEndian, int32(len(payload)))
	buf.Write(payload)

	if _, err := c.pipe.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("write to Discord IPC pipe: %w", err)
	}

	resp := make([]byte, readBufSize)
	n, err := c.pipe.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("read from Discord IPC pipe: %w", err)
	}
	if n < 8 {
		return nil, fmt.Errorf("short response from Discord IPC pipe: %d bytes", n)
	}
	return resp[8:n], nil
}
