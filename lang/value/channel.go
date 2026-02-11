package value

// NOTE: Add to value.go const block:
//   ChannelType  Type = "channel"

// Channel wraps a Go channel for Aiki concurrency.
type Channel struct {
	C chan Value
}

func (c *Channel) Type() Type      { return ChannelType }
func (c *Channel) Inspect() string { return "<channel>" }

func NewChannel() *Channel {
	return &Channel{C: make(chan Value)}
}
