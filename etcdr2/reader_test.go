package etcdr2

import (
	"encoding/json"
	"testing"

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

//go:generate curl http://127.0.0.1:2379/v2/keys/messages4/test -XPUT -d value=Hello
func TestKeyReaderYaml(t *testing.T) {
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
