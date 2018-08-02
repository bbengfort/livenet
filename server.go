package livenet

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bbengfort/livenet/pb"
	"github.com/bbengfort/x/peers"
	"google.golang.org/grpc"
)

// Server implements a LiveNet host that connects to all peers on the network
// via gRPC streams. It can send a variety of messages but primarily sends
// routine heartbeats to the other servers.
type Server struct {
	peers.Peer

	config  *Config    // Configuration of the service
	remotes []*Remote  // Remote peers on the network
	events  chan Event // Event handling channel
	clients uint64     //  Number of connected clients
}

// Listen for messages from peers and clients and run the event loop.
func (s *Server) Listen() error {
	// Create the events channel and ensure it is nilified when exhausted
	s.events = make(chan Event, actorEventBufferSize)
	defer func() { s.events = nil }()

	// Open TCP socket to listen for incoming streams
	addr := s.Endpoint(false)
	sock, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s", addr)
	}
	defer sock.Close()
	info("listening for requests on %s", addr)

	// Initialize and run the gRPC server in its own thread
	srv := grpc.NewServer()
	pb.RegisterLiveNetServer(srv, s)
	go func() {
		if err := srv.Serve(sock); err != nil {
			// Dispatch an error event and stop the server
			s.DispatchError(err, srv)
		}
	}()

	// Send off the heartbeat and status ticker
	go s.Heartbeat()
	go s.Status()

	// Run the event handling loop
	for event := range s.events {
		if err := s.Handle(event); err != nil {
			return err
		}
	}

	return nil
}

// Close the event handler and shutdown the server gracefully.
func (s *Server) Close() error {
	if s.events == nil {
		return errors.New("server is not currently listening for events")
	}

	close(s.events)
	return nil
}

// Post implements the LiveNet stream server, listening for stream connections
// from clients and remote hosts and dispatching each message as an event.
// Every message received on the stream is responded to before a new message
// event is dispatched on receive. This ensures that messages are ordered with
// respect to those that are sent from the client.
//
// Post and Dispatch together also implements the simple multiplexer based on
// message type. Post sends events of the specified type, which gets handled
// by the specific event handler.
func (s *Server) Post(stream pb.LiveNet_PostServer) (err error) {
	var (
		client   string
		messages uint64
		envelope *pb.Envelope
	)

	// Increment the current number of clients and decrement when done
	s.clients++
	defer func() { s.clients-- }()

	// Keep receiving messages on the stream until the client disconnects,
	// send a reply after each message is received and handled by the server.
	for {
		if envelope, err = stream.Recv(); err != nil {
			// Log the disconnected client
			if client == "" {
				info("client disconnected before first message")
			} else {
				info("%s disconnected after %d messages", client, messages)
			}

			// Return the error
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Log the client connection
		if client == "" {
			client = envelope.Sender
			info("%s connected to %s", client, s.Name)
		}

		// Create a channel to wait for the event handler
		source := make(chan *pb.Envelope, 1)

		// Dispatch the message event and wait for it to be handled
		if err = s.DispatchMessage(envelope, source); err != nil {
			return err
		}

		// Wait for the event to be handled before receiving the next message
		// on the stream. This ensures that the order of messages received
		// matches the order of replies sent.
		if err = stream.Send(<-source); err != nil {
			return err
		}

		// Increment the number of handled messages
		messages++
	}
}

// Dispatch an event to be serialized by the event channel.
func (s *Server) Dispatch(e Event) error {
	if s.events == nil {
		return errors.New("server is not currently listening for events")
	}

	s.events <- e
	return nil
}

// DispatchMessage creates an event for the specified message type
func (s *Server) DispatchMessage(msg *pb.Envelope, source interface{}) error {
	return s.Dispatch(&event{etype: MessageEvent, source: source, value: msg})
}

// DispatchError sends error messages that will stop the server and the event
// loop, returning an error and closing the process.
func (s *Server) DispatchError(err error, source interface{}) {
	s.Dispatch(&event{etype: ErrorEvent, source: source, value: err})
}

// Handle events by passing the event to the specified event handlers.
func (s *Server) Handle(e Event) error {
	trace("%s event received: %v", e.Type(), e.Value())

	switch e.Type() {
	case ErrorEvent:
		return e.Value().(error)
	case HeartbeatTimeout:
		return s.onHeartbeatTimeout(e)
	case StatusTimeout:
		return s.onStatusTimeout(e)
	case MessageEvent:
		return s.onMessageEvent(e)
	default:
		return fmt.Errorf("no handler identified for event %s", e.Type())
	}
}
