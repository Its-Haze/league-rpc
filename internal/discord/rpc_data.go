package discord

// RPCData represents the data to be displayed in Discord Rich Presence
// This is equivalent to the Python RPCData dataclass
type RPCData struct {
	LargeImage string // URL to the large image
	LargeText  string // Text shown when hovering over large image
	SmallImage string // URL to the small image
	SmallText  string // Text shown when hovering over small image
	Details    string // First line of text (queue name, etc.)
	State      string // Second line of text (KDA, lobby count, etc.)
	Start      int64  // Unix timestamp for "Elapsed" timer
}

// Equals compares two RPCData instances for equality
// Used to detect if Discord presence actually needs updating
func (r *RPCData) Equals(other *RPCData) bool {
	if other == nil {
		return false
	}

	return r.LargeImage == other.LargeImage &&
		r.LargeText == other.LargeText &&
		r.SmallImage == other.SmallImage &&
		r.SmallText == other.SmallText &&
		r.Details == other.Details &&
		r.State == other.State &&
		r.Start == other.Start
}

// Copy creates a deep copy of the RPCData
func (r *RPCData) Copy() *RPCData {
	if r == nil {
		return nil
	}

	copy := *r
	return &copy
}

// IsEmpty returns true if the RPC data is empty (all fields are zero values)
func (r *RPCData) IsEmpty() bool {
	return r.LargeImage == "" &&
		r.LargeText == "" &&
		r.SmallImage == "" &&
		r.SmallText == "" &&
		r.Details == "" &&
		r.State == "" &&
		r.Start == 0
}
