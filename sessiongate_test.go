package sessiongate

import (
	"crypto/rand"
	"errors"
	"regexp"
	"testing"
	"time"
)

var signKey []byte

func init() {
	signKey = make([]byte, 16)
	rand.Read(signKey)
}

// TestStart tests the START command for the SessionGate module
func TestStart(t *testing.T) {
	t.Run("Should fail with missing SignKey", func(t *testing.T) {
		config := &Config{}

		_, err := NewSessiongate(config)
		if err == nil {
			t.Fail()
		}
	})

	t.Run("Should fail with negative TTL", func(t *testing.T) {
		config := &Config{
			SignKey: signKey,
		}

		sessiongate, err := NewSessiongate(config)
		if err != nil {
			t.Error(err)
		}

		_, err = sessiongate.Start(-1)
		if err == nil {
			t.Error(errors.New("Negative TTL should produce an error"))
		}
	})

	// checkStart checks if the START command does not produce an error and if
	// token is in the expected format
	checkStart := func(config *Config) {
		sessiongate, err := NewSessiongate(config)
		if err != nil {
			t.Error(err)
		}

		token, err := sessiongate.Start(300)
		if err != nil {
			t.Error(err)
		}

		regex := "^v[0-9]\\.[a-zA-Z0-9]+\\.[a-zA-Z0-9]+$"
		match, err := regexp.MatchString(regex, string(token))
		if err != nil {
			t.Error(err)
		}

		if match == false {
			err = errors.New("The response token does not match the expected format")
			t.Error(err)
		}
	}

	t.Run("Should succeed with default configuration", func(t *testing.T) {
		checkStart(&Config{
			SignKey: signKey,
		})
	})

	t.Run("Should succeed with explicit Addr", func(t *testing.T) {
		checkStart(&Config{
			SignKey: signKey,
			Addr:    "localhost:6379",
		})
	})

	t.Run("Should succeed with explicit MaxIdle", func(t *testing.T) {
		checkStart(&Config{
			SignKey: signKey,
			MaxIdle: 15,
		})
	})

	t.Run("Should succeed with explicit IdleTimeout", func(t *testing.T) {
		checkStart(&Config{
			SignKey:     signKey,
			IdleTimeout: 90 * time.Second,
		})
	})
}
