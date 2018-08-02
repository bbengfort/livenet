package livenet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/bbengfort/x/peers"
)

// Default configuration values
const (
	DefaultTick          = 500 * time.Millisecond
	DefaultLogLevel      = LogCaution
	actorEventBufferSize = 1024
)

// Config implements a simple configuration object that can be loaded from a
// JSON file and defines the LiveNet network.
type Config struct {
	Name     string       `json:"name,omitempty"`      // unique name of local replica (hostname by default)
	Seed     int64        `json:"seed"`                // random seed to initialize with
	Tick     string       `json:"tick"`                // click tick rate for timing (parseable duration)
	Uptime   string       `json:"uptime,omitempty"`    // run for a specified duration then shutdown
	LogLevel int          `json:"log_level,omitempty"` // verbosity of logging, lower is more verbose
	Peers    []peers.Peer `json:"peers"`               // all hosts on the LiveNet
}

// Load the configuration from the path on disk
func (c *Config) Load(path string) (err error) {
	var data []byte
	if data, err = ioutil.ReadFile(path); err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

// Dump the configuration to the path on disk
func (c *Config) Dump(path string) (err error) {
	var data []byte
	if data, err = json.Marshal(c); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// GetName returns the configured name or the hostname, required.
func (c *Config) GetName() (string, error) {
	if c.Name != "" {
		return c.Name, nil
	}

	if hostname, err := os.Hostname(); err == nil {
		return hostname, nil
	}

	return "", errors.New("could not find name of local host")
}

// GetPeer returns the local peer configuration or error if no peer found.
func (c *Config) GetPeer() (peers.Peer, error) {
	local, err := c.GetName()
	if err != nil {
		return peers.Peer{}, err
	}

	for _, peer := range c.Peers {
		if peer.Name == local {
			return peer, nil
		}
	}

	return peers.Peer{}, fmt.Errorf("could not find peer for '%s'", local)
}

// GetRemotes returns all peer configurations for remote hosts on the network,
// excluding the local peer configuration.
func (c *Config) GetRemotes(actor Dispatcher) ([]*Remote, error) {
	local, err := c.GetName()
	if err != nil {
		return nil, err
	}

	remotes := make([]*Remote, 0, len(c.Peers)-1)

	for _, peer := range c.Peers {
		if local == peer.Name {
			continue
		}
		remotes = append(remotes, NewRemote(peer, actor))
	}

	return remotes, nil
}

// GetTick returns the parsed duration from the tick configuration
func (c *Config) GetTick() (tick time.Duration, err error) {
	if c.Tick == "" {
		return DefaultTick, nil
	}
	if tick, err = time.ParseDuration(c.Tick); err != nil {
		return tick, fmt.Errorf("could not parse tick: %s", err)
	}
	return tick, nil
}

// GetUptime returns the parsed duration from the uptime configuration. If no
// uptime is specified then a duration of 0 and no error is returned.
func (c *Config) GetUptime() (uptime time.Duration, err error) {
	if c.Uptime == "" {
		return 0, nil
	}
	if uptime, err = time.ParseDuration(c.Uptime); err != nil {
		return uptime, fmt.Errorf("could not parse uptime: %s", err)
	}
	return uptime, nil
}

// GetLogLevel returns the uint8 parsed logging verbosity
func (c *Config) GetLogLevel() uint8 {
	if c.LogLevel > 0 {
		return uint8(c.LogLevel)
	}
	return DefaultLogLevel
}

// SetLogLevel from the configuration if specified, e.g. > 0
func (c *Config) SetLogLevel() {
	SetLogLevel(c.GetLogLevel())
}

// SetSeed if the seed is specified, e.g. > 0
func (c *Config) SetSeed() {
	if c.Seed != 0 {
		debug("setting random seed to %d", c.Seed)
		rand.Seed(c.Seed)
	}
}
