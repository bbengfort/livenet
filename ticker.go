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

	// Sleep for a random interval <= tick so heartbeat messages are sent in
	// the random interval (tick, 2tick) to prevent network saturation.
	sleep := time.Duration(rand.Int63n(int64(tick)))
	time.Sleep(sleep)
}

// Status reports the liveness status to the console.
func (s *Server) Status() {
	tick, err := s.config.GetTick()
	if err != nil {
		return
	}

	// Schedule the next status event when this event is dispatched
	defer time.AfterFunc(tick*1000, s.Status)

	// Dispatch the status event
	s.Dispatch(&event{etype: StatusTimeout, source: nil, value: nil})
}
