package noble

import (
	"encoding/json"
	"errors"
	"strings"
)

// SecretStorage reader interface
type SecretStorage interface {
	Clone() SecretStorage
	Read(path string) (string, error)
}

//nolint:gochecknoglobals
var registered = map[string]SecretStorage{
	"raw":    &rawReader{},
	"env":    &envReader{},
	"dynenv": &dynReader{},
}

type Secret struct {
	source      string
	secrets     []*secret
	single      bool
	parseError  error
	parsedParts []string
}

// secret object
type secret struct {
	reader     SecretStorage
	path       string
	parseError error
	internal   error
}

func (ss *Secret) Error() string {
	if e := ss.ParseError(); e != nil {
		return e.Error()
	}
	if e := ss.InternalError(); e != nil {
		return e.Error()
	}
	return ""
}

// InternalError returns error
func (ss *Secret) InternalError() error {
	if len(ss.secrets) == 0 {
		return ss.parseError
	}
	var err []string
	for _, sr := range ss.secrets {
		if sr.internal != nil {
			err = append(err, sr.internal.Error())
		}
	}
	if err == nil {
		return nil
	}
	return errors.New(strings.Join(err, ";"))
}

// ParseError returns error
func (ss *Secret) ParseError() error {
	var err []string
	for _, sr := range ss.secrets {
		if sr.parseError != nil {
			err = append(err, sr.parseError.Error())
		}
	}
	if err == nil {
		return ss.parseError
	}
	return errors.New(strings.Join(err, ";"))
}

// Register new SecretStorage reader interface
func Register(key string, impl SecretStorage) {
	registered[key] = impl
}

// UnmarshalYAML read secrets from yaml
func (ss *Secret) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string

	if err := unmarshal(&s); err != nil {
		return err
	}
	_ = ss.readAll(s)
	return nil
}

// UnmarshalJSON read secrets from json
func (ss *Secret) UnmarshalJSON(data []byte) error {
	var s string
	if e := json.Unmarshal(data, &s); e != nil {
		return e
	}
	_ = ss.readAll(s)
	return nil
}

// UnmarshalText from text formats
func (sw *secret) UnmarshalText(text []byte) error {
	sw.parseError = sw.read(string(text))
	return nil
}

func (ss *Secret) readAll(in string) error {
	ss.source = in
	if !strings.Contains(in, "{{") {
		ss.single = true
		sr := secret{}.new(in)
		if sr.parseError != nil {
			ss.parseError = sr.parseError
			return ss.parseError
		}
		ss.secrets = []*secret{&sr}
		return nil
	}
	prc := in
	for {
		start := strings.Index(prc, "{{")
		if start == -1 {
			break
		}
		stop := strings.Index(prc, "}}")
		if stop == -1 {
			ss.parseError = errors.New("incorrect format. use [some text]{{<storage>:<path/name>}}[some text]")
			return ss.parseError
		}
		sec := prc[start+2 : stop]
		sr := new(secret)
		if err := sr.read(sec); err != nil {
			ss.parseError = err
		}
		ss.secrets = append(ss.secrets, sr)
		ss.parsedParts = append(ss.parsedParts, prc[:start])
		prc = prc[stop+2:]
	}
	ss.parsedParts = append(ss.parsedParts, prc)
	return nil
}

func (sw *secret) read(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		sw.parseError = errors.New("incorrect format. use <storage>:<path/name>")
		return sw.parseError
	}
	key := parts[0]
	sw.path = parts[1]

	reader, ok := registered[key]
	if !ok {
		sw.parseError = errors.New("unregistered storage: " + key)
		return sw.parseError
	}

	sw.reader = reader.Clone()
	_, sw.internal = sw.reader.Read(sw.path)
	return sw.internal
}

// New static constructor
func (sw secret) new(s string) secret {
	val := secret{}
	_ = val.read(s)
	return val
}

// New static constructor
func (ss Secret) New(s string) Secret {
	val := Secret{}
	_ = val.readAll(s)
	return val
}

// Get value getter
func (sw *secret) Get() string {
	if sw.reader == nil {
		return ""
	}
	val, err := sw.reader.Read(sw.path)
	sw.internal = err
	return val
}

func (ss *Secret) Get() string {
	if len(ss.secrets) == 0 {
		return ss.source
	}
	if ss.single {
		return ss.secrets[0].Get()
	}
	if len(ss.parsedParts) != len(ss.secrets)+1 {
		ss.parseError = errors.New("parser error")
		return ss.source
	}
	s := ss.parsedParts[0]
	for i, sr := range ss.secrets {
		s += sr.Get()
		s += ss.parsedParts[i+1]
	}
	return s
}

type requiredSecretRule struct {
	message string
	skipNil bool
}

// RequiredSecret validation rule
//nolint:gochecknoglobals
var RequiredSecret = &requiredSecretRule{message: "cannot be blank", skipNil: false}

// ToDo
// var OptionalSecret = &requiredSecretRule{message: "cannot be blank", skipNil: true}

func (rd requiredSecretRule) Validate(value interface{}) error {
	s, ok := value.(Secret)
	if !ok {
		return errors.New("invalid type")
	}
	if s.ParseError() != nil {
		return s.ParseError()
	}
	for _, sr := range s.secrets {
		if sr.reader == nil {
			return errors.New("invalid value format. use <storage type>:<path/name/value...(depend on storage type)>")
		}
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
