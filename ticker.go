package livenet

import (
	"math/rand"
	"time"
)

// Heartbeat sends a routine liveness message to other peers.
func (s *Server) Heartbeat() {
	tick, err := s.config.GetTick()
	if err != nil {
		return
	}

	// Schedule the next heartbeat event
	defer time.AfterFunc(tick, s.Heartbeat)

	// Dispatch the heartbeat event
	s.Dispatch(&event{etype: HeartbeatTimeout, source: nil, value: nil})

	// Sleep for a random interval <= half the tick
	sleep := rand.Int63n(int64(tick) / 2)
	time.Sleep(time.Duration(sleep))
}
