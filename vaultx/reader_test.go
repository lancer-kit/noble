package vaultx

// WARNING: non-standard port used in tests (1234 instead 8000)
// For use this test execute `vault kv put  -address="http://127.0.0.1:1234" secret/data pass="my long password"\
// test="passed"`

import (
	"encoding/json"
	"fmt"
	"testing"

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
		SetToken("myroot") //YOUR TOKEN HERE
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