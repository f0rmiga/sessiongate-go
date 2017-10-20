package sessiongate

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

// A Sessiongate represents a connection to the SessionGate module loaded in the
// Redis server.
type Sessiongate struct {
	redisPool *redis.Pool

	signKey []byte
}

// NewSessiongate initializes a new Sessiongate
func NewSessiongate(config *Config) (*Sessiongate, error) {
	// Returns an error if SignKey is not set
	if config.SignKey == nil {
		return nil, errors.New("SignKey is required for the Sessiongate config")
	}

	// Sets addr to a default value if it is an empty string
	var addr string
	if config.Addr == "" {
		addr = ":6379"
	} else {
		addr = config.Addr
	}

	// Sets maxIdle to a default value if it is 0
	var maxIdle int
	if config.MaxIdle == 0 {
		maxIdle = 3
	} else {
		maxIdle = config.MaxIdle
	}

	// Sets idleTimeout to a default value if it is 0
	var idleTimeout time.Duration
	if config.IdleTimeout == 0 {
		idleTimeout = 240 * time.Second
	} else {
		idleTimeout = config.IdleTimeout
	}

	sessiongate := new(Sessiongate)

	// Initialize the Redis connection pool
	sessiongate.redisPool = &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}

	// Sets the SignKey passed to config
	sessiongate.signKey = config.SignKey

	return sessiongate, nil
}

// Start starts a new session in the SessionGate module and returns the
// generated token
func (sessiongate *Sessiongate) Start(ttl int) ([]byte, error) {
	conn := sessiongate.redisPool.Get()
	defer conn.Close()

	r, err := conn.Do("SESSIONGATE.START", sessiongate.signKey, ttl)
	if err != nil {
		return nil, err
	}

	return r.([]byte), nil
}

// Expire sets the TTL for a session in the SessionGate module
func (sessiongate *Sessiongate) Expire(token []byte, ttl int) error {
	conn := sessiongate.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SESSIONGATE.EXPIRE", sessiongate.signKey, token, ttl)
	return err
}

// PSet sets a payload for a session in the SessionGate module
// name is the payload name to be used to get the payload later
func (sessiongate *Sessiongate) PSet(token, name, payload []byte) error {
	conn := sessiongate.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SESSIONGATE.PSET", sessiongate.signKey, token, name, payload)
	return err
}

// PGet gets a payload for a session in the SessionGate module
// name is the payload name
func (sessiongate *Sessiongate) PGet(token, name []byte) ([]byte, error) {
	conn := sessiongate.redisPool.Get()
	defer conn.Close()

	r, err := conn.Do("SESSIONGATE.PGET", sessiongate.signKey, token, name)
	if err != nil {
		return nil, err
	}

	return r.([]byte), nil
}
