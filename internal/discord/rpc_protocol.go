package discord

import (
	"crypto/rand"
	"fmt"
)

// handshake is the opcode-0 payload sent once per Dial to identify the app.
type handshake struct {
	V        string `json:"v"`
	ClientID string `json:"client_id"`
}

// activityFrame is the opcode-1 SET_ACTIVITY payload.
type activityFrame struct {
	Cmd   string       `json:"cmd"`
	Args  activityArgs `json:"args"`
	Nonce string       `json:"nonce"`
}

type activityArgs struct {
	Pid      int              `json:"pid"`
	Activity *activityPayload `json:"activity"`
}

// activityPayload mirrors Discord's Rich Presence activity object, limited
// to the fields RPCData ever sets.
type activityPayload struct {
	Details    string             `json:"details,omitempty"`
	State      string             `json:"state,omitempty"`
	Assets     *activityAssets    `json:"assets,omitempty"`
	Timestamps *activityTimestamp `json:"timestamps,omitempty"`
}

type activityAssets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type activityTimestamp struct {
	Start int64 `json:"start,omitempty"`
}

// newNonce generates a random UUIDv4-shaped nonce; Discord doesn't validate
// its format, just that requests carry one.
func newNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:]), nil
}
