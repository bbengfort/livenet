package livenet

import (
	"context"
	"fmt"
	"sync"

	"github.com/bbengfort/livenet/pb"
	"github.com/bbengfort/x/peers"
	"google.golang.org/grpc"
)

// Remote implements a streaming connection to a remote peer on the network.
type Remote struct {
	sync.RWMutex
	peers.Peer

	actor  Dispatcher            // the listener to dispatch events to
	conn   *grpc.ClientConn      // grpc dial connection to the remote
	client pb.LiveNetClient      // rpc client specified by protobuf
	stream pb.LiveNet_PostClient // message stream to send on
	online bool                  // if the client is connected or not
}

// NewRemote creates a new remote associated with the actor
func NewRemote(p peers.Peer, a Dispatcher) *Remote {
	return &Remote{Peer: p, actor: a}
}

// Send a message to the remote
func (r *Remote) Send(msg *pb.Envelope) error {

	// TODO: how to reconcile the multiple threads trying to send/reconnect
	// the stream with the listener that has better knowledge about the state
	// of the stream? Right now I'm using RLocks to send on the stream.

	// Does not reconnect if already online uses double-checked lock for safety
	r.connect()

	// However, at this point, the recv routine may be closing the connection!
	if err := r.stream.Send(msg); err != nil {
		// go offline because of the error
		caution("dropped message to %s: %s", r.Name, err)
		r.close()
	}

	return nil
}

// Recv messages from the remote and dispatch message received events
func (r *Remote) recv() {
	var (
		msg *pb.Envelope
		err error
	)

	for {
		if msg, err = r.stream.Recv(); err != nil {
			// If we can no longer receive from the stream, close the conn
			caution("stream to %s closed: %s", r.Name, err)
			r.close()
			return
		}

		// Dispatch the received message event
		r.actor.DispatchMessage(msg, r)
	}

}

//===========================================================================
// Connection Handlers
//===========================================================================

// Connect to the remote and create a stream message stream to it.
func (r *Remote) connect() (err error) {
	// Double-checked locking using RWMutex for safety
	r.RLock()
	if !r.isConnected() {
		r.RUnlock()
		r.Lock()
		defer r.Unlock()

		addr := r.Endpoint(false)

		if r.conn, err = grpc.Dial(addr, grpc.WithInsecure()); err != nil {
			return fmt.Errorf("could not connect to '%s': %s", addr, err)
		}

		r.client = pb.NewLiveNetClient(r.conn)
		if r.stream, err = r.client.Post(context.Background()); err != nil {
			r.close() // reset the state of the remote
			return fmt.Errorf("could not create message stream to '%s': %s", addr, err)
		}

		// At this point we can say we are connected because the stream is good
		r.toggleOnline(true)

		// Run the go routine that handles replies and dispatches reply events
		go r.recv()
		return nil

	}
	r.RUnlock()
	return nil
}

// Close the connection to the remote and cleanup the client
func (r *Remote) close() (err error) {
	// Protect the connection
	r.Lock()
	defer r.Unlock()

	// Ensure valid state after close
	defer func() {
		r.conn = nil
		r.client = nil
		r.stream = nil
		r.toggleOnline(false)
	}()

	if r.stream != nil {
		if err = r.stream.CloseSend(); err != nil {
			return fmt.Errorf("could not close stream to %s: %s", r.Name, err)
		}
	}

	if r.conn != nil {
		if err = r.conn.Close(); err != nil {
			return fmt.Errorf("could not close connection to %s: %s", r.Name, err)
		}
	}

	return nil
}

// isConnected returns true if the connection is not nil (not thread-safe)
func (r *Remote) isConnected() bool {
	return r.conn != nil && r.stream != nil
}

// Set the online state and issue a message if the state has changed (not thread-safe)
func (r *Remote) toggleOnline(online bool) {
	if online && !r.online {
		info("connection to %s (%s) is now online", r.Name, r.Endpoint(false))
	} else if !online && r.online {
		info("disconnected from %s (%s)", r.Name, r.Endpoint(false))
	}

	r.online = online
}
