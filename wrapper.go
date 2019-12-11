package noble

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// SecretStorage reader interface
type SecretStorage interface {
	Read(path string) (string, error)
}

var registered = map[string]SecretStorage{
	"raw":    rawReader{}.Interface(),
	"env":    envReader{}.Interface(),
	"dynenv": dynReader{}.Interface(),
}

// Secret object
type Secret struct {
	reader   SecretStorage
	path     string
	internal error
}

func (sw Secret) Error() string {
	if sw.internal != nil {
		return sw.internal.Error()
	}
	return ""
}

// InternalError returns error
func (sw Secret) InternalError() error {
	return sw.internal
}

// Register new SecretStorage reader interface
func Register(key string, impl SecretStorage) {
	registered[key] = impl
}

// UnmarshalYAML read secrets from yaml
func (sw *Secret) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string

	if err := unmarshal(&s); err != nil {
		return err
	}
	return sw.read(s)
}

// UnmarshalJSON read secrets from json
func (sw *Secret) UnmarshalJSON(data []byte) error {
	var s string
	if e := json.Unmarshal(data, &s); e != nil {
		return e
	}
	return sw.read(s)
}

func (sw *Secret) read(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		sw.internal = errors.New("incorrect format. use <storage>:<path/name>")
		return sw.internal
	}
	key := parts[0]
	sw.path = parts[1]
	var ok bool
	if sw.reader, ok = registered[key]; !ok {
		sw.internal = errors.New("unregistered storage: " + key)
		return sw.internal
	}
	_, sw.internal = sw.reader.Read(sw.path)
	return sw.internal
}

// New static constructor
func (sw Secret) New(s string) Secret {
	val := Secret{}
	_ = val.read(s)
	return val
}

// Get value getter
func (sw *Secret) Get() string {
	val, err := sw.reader.Read(sw.path)
	sw.internal = err
	return val
}

type rawReader struct{}

func (rr rawReader) Read(path string) (string, error) {
	return path, nil
}

func (rr rawReader) Interface() SecretStorage {
	return &rawReader{}
}

// Read parameter value from environment variable and store into "cache"
type envReader struct {
	cached string
}

//Read env.variable into internal cache
func (er *envReader) Read(path string) (string, error) {
	if er.cached == "" {
		er.cached = os.Getenv(path)
	}
	if er.cached == "" {
		return er.cached, errors.New("unable to read OS environment variable:" + path)
	}
	return er.cached, nil
}

// Interface constructor for envReader
func (er envReader) Interface() SecretStorage {
	return &envReader{}
}

// Read parameter value from environment variable dynamically
type dynReader struct {
}

//Read env.variable dynamically
func (d *dynReader) Read(path string) (string, error) {
	val := os.Getenv(path)
	if val == "" {
		return val, errors.New("unable to read OS environment variable:" + path)
	}
	return val, nil
}

// Interface constructor for dynReader
func (d dynReader) Interface() SecretStorage {
	return &dynReader{}
}
