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
	reader     SecretStorage
	path       string
	parseError error
	internal   error
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

// InternalError returns error
func (sw Secret) ParseError() error {
	return sw.parseError
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
	sw.parseError = sw.read(s)
	return nil
}

// UnmarshalJSON read secrets from json
func (sw *Secret) UnmarshalJSON(data []byte) error {
	var s string
	if e := json.Unmarshal(data, &s); e != nil {
		return e
	}
	sw.parseError = sw.read(s)
	return nil
}

func (sw *Secret) read(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		sw.parseError = errors.New("incorrect format. use <storage>:<path/name>")
		return sw.parseError
	}
	key := parts[0]
	sw.path = parts[1]
	var ok bool
	if sw.reader, ok = registered[key]; !ok {
		sw.parseError = errors.New("unregistered storage: " + key)
		return sw.parseError
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
	if sw.reader == nil {
		return ""
	}
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

type requiredSecretRule struct {
	message string
	skipNil bool
}

var RequiredSecret = &requiredSecretRule{message: "cannot be blank", skipNil: false}

func (rd requiredSecretRule) Validate(value interface{}) error {
	s, ok := value.(Secret)
	if !ok {
		return errors.New("invalid type")
	}
	if s.ParseError() != nil {
		return s.ParseError()
	}
	if s.reader == nil {
		return errors.New("invalid value format. use <storage type>:<path/name/value...(depend on storage type)>")
	}
	s.Get()
	return s.InternalError()
}

func (rd *requiredSecretRule) Error(message string) *requiredSecretRule {
	msg := rd.message
	if msg == "" {
		msg = message
	} else {
		msg += ": " + message
	}
	return &requiredSecretRule{
		message: msg,
		skipNil: rd.skipNil,
	}
}
