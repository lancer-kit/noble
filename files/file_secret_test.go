package files

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/lancer-kit/noble"
)

const (
	testValue = "package files"
	testYaml  = `
secret: "file:./file_secret_test.go"
`
	testJson = `
{
"secret": "file:./file_secret_test.go"
}
`
)

type testConfig struct {
	Secret noble.Secret `yaml:"secret"`
}

func TestFileReaderYaml(t *testing.T) {
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
