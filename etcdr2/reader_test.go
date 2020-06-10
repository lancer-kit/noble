package etcdr2

import (
	"encoding/json"
	"testing"

	"github.com/lancer-kit/armory/api/httpx"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/lancer-kit/noble"
)

const (
	testValue = "Hello"
	testYaml  = `
secret: "etcd2:messages4/test"
`
	testJson = `
{
"secret": "etcd2:messages4/test"
}
`
)

type testConfig struct {
	Secret noble.Secret `yaml:"secret"`
}

func installedETCD() bool {
	resp, err := httpx.NewXClient().Get("http://localhost:2379/metrics")
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	_ = resp.Body.Close()
	return true
}

//go:generate curl http://127.0.0.1:2379/v2/keys/messages4/test -XPUT -d value=Hello
func TestKeyReaderYaml(t *testing.T) {
	if !installedETCD() {
		println("no etcd installed")
		return
	}
	var c testConfig
	e := yaml.Unmarshal([]byte(testYaml), &c)
	assert.NoError(t, e)
	assert.NoError(t, c.Secret.InternalError())
	assert.NoError(t, c.Secret.ParseError())
	if !assert.Equal(t, testValue, c.Secret.Get()) {
		return
	}
	println("value:", c.Secret.Get())
}

func TestFileReaderJson(t *testing.T) {
	if !installedETCD() {
		println("no etcd installed")
		return
	}
	var c testConfig
	e := json.Unmarshal([]byte(testJson), &c)
	assert.NoError(t, e)
	assert.NoError(t, c.Secret.InternalError())
	assert.NoError(t, c.Secret.ParseError())
	if !assert.Equal(t, testValue, c.Secret.Get()) {
		return
	}
	println("value:", c.Secret.Get())
}
