package sessiongate

import (
	"crypto/rand"
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"
)

var signKey []byte

func init() {
	signKey = make([]byte, 16)
	rand.Read(signKey)
}

// TestInitializer tests the Sessiongate initializer
func TestInitializer(t *testing.T) {
	t.Run("Should fail with missing SignKey", func(t *testing.T) {
		config := &Config{}

		_, err := NewSessiongate(config)
		if err == nil {
			t.Fail()
		}
	})
}

// TestStart tests the START command for the SessionGate module
func TestStart(t *testing.T) {
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

func createSession() (*Sessiongate, []byte, error) {
	config := &Config{
		SignKey: signKey,
	}

	sessiongate, err := NewSessiongate(config)
	if err != nil {
		return nil, nil, err
	}

	token, err := sessiongate.Start(300)
	if err != nil {
		return nil, nil, err
	}

	return sessiongate, token, nil
}

// TestExpire tests the EXPIRE command for the SessionGate module
func TestExpire(t *testing.T) {
	sessiongate, token, err := createSession()
	if err != nil {
		t.Error(err)
	}

	t.Run("Should fail with negative TTL", func(t *testing.T) {
		err := sessiongate.Expire(token, -5)
		if err == nil {
			t.Error(errors.New("Negative TTL should produce an error"))
		}
	})

	t.Run("Should succeed with positive TTL", func(t *testing.T) {
		err := sessiongate.Expire(token, 500)
		if err != nil {
			t.Error(err)
		}
	})
}

// TestPSet tests the PSET command for the SessionGate module
func TestPSet(t *testing.T) {
	sessiongate, token, err := createSession()
	if err != nil {
		t.Error(err)
	}

	t.Run("Should fail with an empty name", func(t *testing.T) {
		name := []byte("")
		payload := []byte("{}")
		err := sessiongate.PSet(token, name, payload)
		if err == nil {
			t.Error(errors.New("Empty name should produce an error"))
		}
	})

	t.Run("Should fail with an empty payload", func(t *testing.T) {
		name := []byte("user")
		payload := []byte("")
		err := sessiongate.PSet(token, name, payload)
		if err == nil {
			t.Error(errors.New("Empty payload should produce an error"))
		}
	})

	t.Run("Should succeed with a JSON string as a payload", func(t *testing.T) {
		name := []byte("user")
		payload := []byte("{\"name\":\"John Doe\"}")
		err := sessiongate.PSet(token, name, payload)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Should succeed with random bytes in the name", func(t *testing.T) {
		name := make([]byte, 8)
		rand.Read(name)
		payload := []byte("{\"name\":\"John Doe\"}")
		err := sessiongate.PSet(token, name, payload)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Should succeed with random bytes in the payload", func(t *testing.T) {
		name := []byte("user")
		payload := make([]byte, 128)
		rand.Read(payload)
		err := sessiongate.PSet(token, name, payload)
		if err != nil {
			t.Error(err)
		}
	})
}

// TestPGet tests the PGET command for the SessionGate module
func TestPGet(t *testing.T) {
	sessiongate, token, err := createSession()
	if err != nil {
		t.Error(err)
	}

	t.Run("Should fail with an empty name", func(t *testing.T) {
		name := []byte("")
		_, err := sessiongate.PGet(token, name)
		if err == nil {
			t.Error(errors.New("Empty name should produce an error"))
		}
	})

	t.Run("Should succeed with a JSON string as a payload", func(t *testing.T) {
		name := []byte("user")
		payload := []byte("{\"name\":\"John Doe\"}")
		err := sessiongate.PSet(token, name, payload)
		if err != nil {
			t.Error(err)
		}

		payloadPGet, err := sessiongate.PGet(token, name)
		if err != nil {
			t.Error(err)
		}

		if reflect.DeepEqual(payload, payloadPGet) == false {
			t.Error(errors.New("The payloads should be equal"))
		}
	})

	t.Run("Should succeed with random bytes in the payload", func(t *testing.T) {
		name := []byte("user")
		payload := make([]byte, 128)
		rand.Read(payload)
		err := sessiongate.PSet(token, name, payload)
		if err != nil {
			t.Error(err)
		}

		payloadPGet, err := sessiongate.PGet(token, name)
		if err != nil {
			t.Error(err)
		}

		if reflect.DeepEqual(payload, payloadPGet) == false {
			t.Error(errors.New("The payloads should be equal"))
		}
	})

	t.Run("Should succeed with random bytes in the name", func(t *testing.T) {
		name := make([]byte, 8)
		rand.Read(name)
		payload := []byte("{\"name\":\"John Doe\"}")
		err := sessiongate.PSet(token, name, payload)
		if err != nil {
			t.Error(err)
		}

		payloadPGet, err := sessiongate.PGet(token, name)
		if err != nil {
			t.Error(err)
		}

		if reflect.DeepEqual(payload, payloadPGet) == false {
			t.Error(errors.New("The payloads should be equal"))
		}
	})
}
