package livenet

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bbengfort/x/noplog"
	"google.golang.org/grpc/grpclog"
)

// PackageVersion of LiveNet
const PackageVersion = "0.2"

// Initialize the package and random numbers, etc.
func init() {
	rand.Seed(time.Now().UnixNano())

	// Initialize our debug logging with our prefix
	SetLogger(log.New(os.Stdout, "[livenet] ", log.Lmicroseconds))
	cautionCounter = new(counter)
	cautionCounter.init()

	// Stop the grpc verbose logging
	grpclog.SetLogger(noplog.New())
}

// New is the entry point to the LiveNet service for a single machine, it
// instantiates a LiveNet sever for the specified network and configuration.
func New(config *Config) (server *Server, err error) {
	// Check the tick and the uptime
	if _, err = config.GetTick(); err != nil {
		return nil, err
	}

	if _, err = config.GetUptime(); err != nil {
		return nil, err
	}

	// Set the logging level and the random seed
	config.SetLogLevel()
	config.SetSeed()

	// Create the server object
	server = &Server{config: config}
	if server.Peer, err = config.GetPeer(); err != nil {
		return nil, err
	}

	// Create the remotes
	if server.remotes, err = config.GetRemotes(server); err != nil {
		return nil, err
	}

	return server, nil
}
