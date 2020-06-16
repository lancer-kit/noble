// Package vaultx Noble vault reader
package vaultx

import (
	"errors"
	"os"
	"strings"

	"github.com/lancer-kit/noble"
	"github.com/sirupsen/logrus"
)

// InitVault required for using "vault:<key>"
func InitVault(cfg *VaultCfg) error {
	var err error
	if cfg == nil {
		cfg = &defaultConfig
	}
	vs, err = newStorage(*cfg)
	return err
}

// SetLogger set logrus entry as log source
func SetLogger(l *logrus.Entry) bool {
	if vs == nil {
		return false
	}
	if l == nil {
		return false
	}
	vs.SetLogger(l)
	return true
}

// SetServerAddress set vault server address
func SetServerAddress(addr string) {
	defaultConfig.ServerAddress = addr
}

// SetSecretPath set vault k/v path. Used secret/data by default
func SetSecretPath(path string) {
	if path[len(path)-1:] == "/" {
		path = path[:len(path)-1]
	}
	defaultConfig.SecretPath = path
}

// SetToken set vault token to login
func SetToken(token string) {
	defaultConfig.Token = token
}

// SetTokenEnv set vault token to login from environment var
func SetTokenEnv(envName string) bool {
	var ok bool
	defaultConfig.Token, ok = os.LookupEnv(envName)
	return ok
}

//SetTokenTTL token time to live in hours
func SetTokenTTL(ttl int64) {
	defaultConfig.TokenTTLHours = ttl
}

//nolint:gochecknoglobals
var vs storage

//nolint:gochecknoinits
func init() {
	noble.Register("vault", &KeyReader{})
}

//KeyReader type implements noble.SecretStorage
type KeyReader struct {
	//key string
}

func (r *KeyReader) Read(key string) (string, error) {
	if vs == nil {
		return "", errors.New("vault connection not initialized")
	}
	parts := strings.Split(key, "?")
	if len(parts) != 2 {
		return "", errors.New("incorrect key format. use \"/<path>?<key>\"")
	}
	if parts[0][:1] != "/" {
		parts[0] = "/" + parts[0]
	}
	return vs.Get(parts[0], parts[1])
}

// Clone returns new empty instance of KeyReader
func (r *KeyReader) Clone() noble.SecretStorage {
	return &KeyReader{}
}
