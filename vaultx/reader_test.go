package vaultx

// WARNING: non-standard port used in tests (1234 instead 8000)
// For use this test execute `vault kv put  -address="http://127.0.0.1:1234" secret/data pass="my long password"\
// test="passed"`

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/lancer-kit/armory/api/httpx"

	"github.com/lancer-kit/noble"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	testYaml = `
secret: "vault:/data?pass"
`

	testJSON = `
{
"secret": "vault:/data?pass"
}
`
)

type testConfig struct {
	Secret noble.Secret `yaml:"secret"`
}

//go:generate curl http://127.0.0.1:1234/v1/sys/health
func installedVault() bool {
	resp, err := httpx.NewXClient().Get("http://127.0.0.1:1234/v1/sys/health")
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	_ = resp.Body.Close()
	return true
}

func initStorage(t *testing.T) {
	if vs != nil {
		return
	}
	SetServerAddress("http://127.0.0.1:1234")
	if !SetTokenEnv("VAULT_TOKEN") {
		SetToken("myroot") // YOUR TOKEN HERE
	}
	err := InitVault(nil)
	assert.NoError(t, err)
	assert.Equal(t, false, SetLogger(nil))
}

func TestKeyReader_Read(t *testing.T) {
	if !installedVault() {
		println("no vault installed and run on port 1234")
		return
	}
	initStorage(t)
	r := KeyReader{}
	v, e := r.Read("/data?test")
	assert.NoError(t, e)
	assert.Equal(t, "passed", v)
}

func TestKeyReader_YAML(t *testing.T) {
	if !installedVault() {
		println("no vault installed and run on port 1234")
		return
	}

	initStorage(t)
	var c testConfig
	e := yaml.Unmarshal([]byte(testYaml), &c)
	assert.NoError(t, e)
	assert.NoError(t, c.Secret.InternalError())
	assert.NoError(t, c.Secret.ParseError())
	x := c.Secret.Get()
	fmt.Printf("%+v\n", x)
}

func TestKeyReader_JSON(t *testing.T) {
	if !installedVault() {
		println("no vault installed and run on port 1234")
		return
	}

	initStorage(t)
	var c testConfig
	e := json.Unmarshal([]byte(testJSON), &c)
	assert.NoError(t, e)
	assert.NoError(t, c.Secret.InternalError())
	assert.NoError(t, c.Secret.ParseError())
	x := c.Secret.Get()
	fmt.Printf("%+v\n", x)
}

func TestSetSecretPath(t *testing.T) {
	SetSecretPath("/some/vault/data/")
	assert.Equal(t, "/some/vault/data", defaultConfig.SecretPath)
}

func TestInitVault(t *testing.T) {
	if !installedVault() {
		assert.Error(t, InitVault(nil))
	} else {
		assert.NoError(t, InitVault(nil))
	}
}

func TestKeyReader_Clone(t *testing.T) {
	m := &KeyReader{}
	r := m.Clone()
	assert.EqualValues(t, m, r)
}

func TestKeyReader_Read1(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		mock    storage
		want    string
		wantErr bool
	}{
		{
			name:    "error",
			args:    args{key: "/some?key"},
			mock:    nil,
			want:    "",
			wantErr: true,
		},
		{
			name: "no error",
			args: args{key: "/some?key"},
			mock: &storageMock{
				data: "return",
				err:  nil,
			},
			want:    "return",
			wantErr: false,
		},
		{
			name: "error key",
			args: args{key: "some"},
			mock: &storageMock{
				err: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &KeyReader{}
			vs = tt.mock
			got, err := r.Read(tt.args.key)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			if got != tt.want {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetLogger(t *testing.T) {
	vs = nil
	assert.False(t, SetLogger(logrus.WithField("unit", "test")))
	vs = &storageMock{}
	assert.True(t, SetLogger(logrus.WithField("unit", "test")))
}

func TestSetSecretPath1(t *testing.T) {
	const n = "/secret/data/some-set"
	SetSecretPath(n)
	assert.Equal(t, n, defaultConfig.SecretPath)
}

func TestSetServerAddress(t *testing.T) {
	const n = "localhost:1000"
	SetServerAddress(n)
	assert.Equal(t, n, defaultConfig.ServerAddress)
}

func TestSetToken(t *testing.T) {
	const n = "token"
	SetToken(n)
	assert.Equal(t, n, defaultConfig.Token)
}

func TestSetTokenEnv(t *testing.T) {
	const n = "SOME_NOT_EXISTING"
	assert.False(t, SetTokenEnv(n))
	assert.NoError(t, os.Setenv(n, n))
	assert.True(t, SetTokenEnv(n))
	assert.Equal(t, n, defaultConfig.Token)
	assert.NoError(t, os.Unsetenv(n))
}

func TestSetTokenTTL(t *testing.T) {
	const n int64 = 2
	SetTokenTTL(n)
	assert.Equal(t, n, defaultConfig.TokenTTLHours)
}

type storageMock struct {
	data string
	err  error
	log  *logrus.Entry
}

func (s *storageMock) Get(_, _ string) (string, error) {
	return s.data, s.err
}

func (s *storageMock) SetLogger(entry *logrus.Entry) {
	s.log = entry
}
