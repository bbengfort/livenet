# LiveNet

**Demo of gRPC streaming for liveness detection in a fully connected network.**

Proof of concept for a fully connected network that uses gRPC streams for constant messaging without exchanging meta information between RPC requests. By detecting if the stream is closed, each host on the network can determine if the connection is live or not. Hosts can reconnect at anytime to repair the state of the network.

This system also demonstrates the use of a message multiplexer so that a single stream between hosts is used. This model hijacks the gRPC server handler definitions, but the hosts implement an actor model so each message is treated as an independent event.

![LiveNet Architecture](fixtures/livenet.png)
