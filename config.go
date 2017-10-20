package sessiongate

import "time"

// Config represents a configuration passed to the Sessiongate initializer
type Config struct {
	SignKey []byte

	Addr        string        // The Redis address
	MaxIdle     int           // The max idle time for a Redis connection
	IdleTimeout time.Duration // The idle timeout for a Redis connection
}
