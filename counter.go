package livenet

import "fmt"

// MessageCounts is a simple data structure for keeping track of how many
// messages are sent, received, and dropped from a connection.
type MessageCounts struct {
	sent uint64
	recv uint64
	drop uint64
}

// Sent increments the sent messages count
func (c *MessageCounts) Sent() {
	c.sent++
}

// Recv increments the received messages count
func (c *MessageCounts) Recv() {
	c.recv++
}

// Drop increments the dropped messages count
func (c *MessageCounts) Drop() {
	c.drop++
}

func (c *MessageCounts) String() string {
	return fmt.Sprintf(
		"%d messages sent, %d recieved (%0.2f%%), %d dropped (%0.2f%%)",
		c.sent, c.recv, c.RecvR(), c.drop, c.DropR(),
	)
}

// RecvR returns the ratio of received to sent messages
func (c *MessageCounts) RecvR() float64 {
	if c.sent == 0 || c.recv == 0 {
		return 0.0
	}
	return float64(c.recv) / float64(c.sent)
}

// DropR returns the ratio of dropped to sent messages
func (c *MessageCounts) DropR() float64 {
	if c.sent == 0 || c.drop == 0 {
		return 0.0
	}
	return float64(c.drop) / float64(c.sent)
}

// Reset the message counts back to zero
func (c *MessageCounts) Reset() {
	c.sent = 0
	c.drop = 0
	c.recv = 0
}
