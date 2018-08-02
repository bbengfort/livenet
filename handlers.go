package livenet

import "github.com/bbengfort/livenet/pb"

// Broadcast a heartbeat message to all remote peers.
func (s *Server) onHeartbeatTimeout(e Event) error {
	debug("heartbeat timeout")

	msg := pb.Wrap(s.Name, pb.MessageType_HEARTBEAT, nil)
	for _, remote := range s.remotes {
		if err := remote.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

// Send acknowledgment to the heartbeat message
func (s *Server) onMessageEvent(e Event) error {
	in := e.Value().(*pb.Envelope)
	debug("received %s message from %s", in.Type, in.Sender)

	msg := pb.Wrap(s.Name, pb.MessageType_HEARTBEAT, nil)

	// TODO: better message type handling rather than checking the source
	if src, ok := e.Source().(chan *pb.Envelope); ok {
		src <- msg
	}
	return nil
}
