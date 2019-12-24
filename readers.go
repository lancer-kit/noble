package noble

import (
	"errors"
	"os"
)

type rawReader struct{}

func (rr rawReader) Read(path string) (string, error) {
	return path, nil
}

// Clone returns new empty instance of rawReader
func (rr rawReader) Clone() SecretStorage {
	return &rawReader{}
}

// Read parameter value from environment variable and store into "cache"
type envReader struct {
	cached string
}

// Read env.variable into internal cache
func (er *envReader) Read(path string) (string, error) {
	if er.cached == "" {
		er.cached = os.Getenv(path)
	}
	if er.cached == "" {
		return er.cached, errors.New("unable to read OS environment variable:" + path)
	}
	return er.cached, nil
}

// Clone returns new empty instance of envReader
func (er envReader) Clone() SecretStorage {
	return &envReader{}
}

// Read parameter value from environment variable dynamically
type dynReader struct{}

// Read env.variable dynamically
func (d *dynReader) Read(path string) (string, error) {
	val := os.Getenv(path)
	if val == "" {
		return val, errors.New("unable to read OS environment variable:" + path)
	}
	return val, nil
}

// Clone returns new empty instance of dynReader
func (d dynReader) Clone() SecretStorage {
	return &dynReader{}
}
